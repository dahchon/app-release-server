package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/dahchon/app-release-server/common"
	"github.com/dahchon/app-release-server/db"
	"github.com/gin-gonic/gin"
)

var (
	FILE_STORAGE_PATH = os.Getenv("ARS_FILE_STORAGE_PATH")
)

type AppDetails struct {
	AppVersion string `json:"app_version"`
	AppBuild   string `json:"app_build"`
	AppName    string `json:"app_name"`
	GitCommit  string `json:"git_commit,omitempty"`
}

func main() {
	listenPort := flag.String("port", "8080", "Listen port")
	flag.Parse()

	if FILE_STORAGE_PATH == "" {
		log.Fatal("You must provide a file storage path")
	}

	r := gin.Default()

	client := db.NewClient()
	connetErr := client.Connect()
	if connetErr != nil {
		log.Fatal(connetErr)
	}
	prismaCtx := context.Background()

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "password", // replace with your own user and password
	}))

	authorized.POST("/admin/upload/app", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		detailsStr := c.PostForm("details")
		var appDetails AppDetails
		if err := json.Unmarshal([]byte(detailsStr), &appDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if appDetails.AppName == "" || appDetails.AppVersion == "" || appDetails.AppBuild == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app name, version or build"})
			return
		}

		dir := fmt.Sprintf("%s/%s/%s/%s", FILE_STORAGE_PATH, appDetails.AppName, appDetails.AppVersion, appDetails.AppBuild)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fileName := filepath.Base(file.Filename)
		dst := fmt.Sprintf("%s/%s", dir, fileName)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		release, err := client.AppRelease.CreateOne(
			db.AppRelease.AppName.Set(appDetails.AppName),
			db.AppRelease.AppVersion.Set(appDetails.AppVersion),
			db.AppRelease.AppBuild.Set(appDetails.AppBuild),
			db.AppRelease.GitCommit.Set(appDetails.GitCommit),
			db.AppRelease.MainFileName.Set(fileName),
		).Exec(prismaCtx)

		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s uploaded successfully", file.Filename), "release": release})
	})

	r.GET("/apps/:app_name/:app_version/:app_build/:file_name", func(c *gin.Context) {
		appName := c.Param("app_name")
		appVersion := c.Param("app_version")
		appBuild := c.Param("app_build")
		fileName := c.Param("file_name")

		if appName == "" || appVersion == "" || appBuild == "" || fileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters: app name, version, build or file name"})
			return
		}

		filePath := fmt.Sprintf("%s/%s/%s/%s/%s", FILE_STORAGE_PATH, appName, appVersion, appBuild, fileName)
		c.File(filePath)
	})

	r.GET("/apps/:app_name/latest", func(c *gin.Context) {
		// get the metadata for the latest release
		appName := c.Param("app_name")
		if appName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameter: app name"})
			return
		}

		release, err := client.AppRelease.FindFirst(
			db.AppRelease.AppName.Equals(appName),
		).OrderBy(
			db.AppRelease.CreatedAt.Order(db.SortOrderDesc),
		).Exec(prismaCtx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if release == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No release found"})
			return
		}

		latest_model := common.AppLatestModel{
			AppVersion:  release.AppVersion,
			AppBuild:    release.AppBuild,
			AppName:     release.AppName,
			DownloadURL: fmt.Sprintf("/apps/%s/%s/%s/%s", release.AppName, release.AppVersion, release.AppBuild, release.MainFileName),
		}

		c.JSON(http.StatusOK, latest_model)
	})

	r.Run(fmt.Sprintf(":%s", *listenPort))
}

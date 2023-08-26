package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/dahchon/app-release-server/db"
	"github.com/gin-gonic/gin"
)

type AppDetails struct {
	AppVersion string `json:"app_version"`
	AppBuild   string `json:"app_build"`
	AppName    string `json:"app_name"`
	GitCommit  string `json:"git_commit,omitempty"`
}

func main() {
	r := gin.Default()

	client := db.NewClient()
	ctx := context.Background()

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "password", // replace with your own user and password
	}))

	authorized.POST("/admin/upload/app", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var json AppDetails
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if json.AppName == "" || json.AppVersion == "" || json.AppBuild == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app name, version or build"})
			return
		}

		dir := fmt.Sprintf("./apps/%s/%s/%s", json.AppName, json.AppVersion, json.AppBuild)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dst := fmt.Sprintf("%s/binary.app", dir)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		release, err := client.AppRelease.CreateOne(
			db.AppRelease.AppName.Set(json.AppName),
			db.AppRelease.AppVersion.Set(json.AppVersion),
			db.AppRelease.AppBuild.Set(json.AppBuild),
			db.AppRelease.GitCommit.Set(json.GitCommit),
		).Exec(c)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("File %s uploaded successfully", file.Filename), "release": release})
	})

	r.GET("/apps/:app_name/:app_version/:app_build/binary.app", func(c *gin.Context) {
		appName := c.Param("app_name")
		appVersion := c.Param("app_version")
		appBuild := c.Param("app_build")

		if appName == "" || appVersion == "" || appBuild == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing app name, version or build"})
			return
		}

		filePath := fmt.Sprintf("./apps/%s/%s/%s/binary.app", appName, appVersion, appBuild)
		c.File(filePath)
	})

	r.Run(":8080")
}

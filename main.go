package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"encoding/json"

	"github.com/dahchon/app-release-server/common"
	"github.com/dahchon/app-release-server/prisma/db"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// print user name and password length
	log.Println("Length of backend username:", len(common.GetBackendUsername()))
	log.Println("Length of backend password:", len(common.GetBackendPassword()))

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	client := db.NewClient()
	connectErr := client.Connect()
	if connectErr != nil {
		log.Fatal(connectErr)
	}

	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == common.GetBackendUsername() && password == common.GetBackendPassword() {
			return true, nil
		}
		return false, nil
	}))

	adminGroup.POST("/upload/app", func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		detailsStr := c.FormValue("details")
		var appDetails AppDetails
		if err := json.Unmarshal([]byte(detailsStr), &appDetails); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		if appDetails.AppName == "" || appDetails.AppVersion == "" || appDetails.AppBuild == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing app name, version or build"})
		}

		dir := fmt.Sprintf("%s/%s/%s/%s", FILE_STORAGE_PATH, appDetails.AppName, appDetails.AppVersion, appDetails.AppBuild)
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		fileName := filepath.Base(file.Filename)
		dst := fmt.Sprintf("%s/%s", dir, fileName)
		// if err := c.SaveUploadedFile(file, dst); err != nil {
		// 	return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		// }
		// echo framework save uploaded file to dst
		// Destination

		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		opendFile, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer opendFile.Close()

		// Copy
		if _, err = io.Copy(opendFile, src); err != nil {
			return err
		}

		ip := c.RealIP()
		release, err := client.AppRelease.CreateOne(
			db.AppRelease.AppName.Set(appDetails.AppName),
			db.AppRelease.AppVersion.Set(appDetails.AppVersion),
			db.AppRelease.AppBuild.Set(appDetails.AppBuild),
			db.AppRelease.GitCommit.Set(appDetails.GitCommit),
			db.AppRelease.MainFileName.Set(fileName),
			db.AppRelease.UploaderIP.Set(ip),
		).Exec(c.Request().Context())

		if err != nil {
			log.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"message": fmt.Sprintf("File %s uploaded successfully", file.Filename), "release": release})
	})

	e.GET("/apps/:app_name/:app_version/:app_build/:file_name", func(c echo.Context) error {
		appName := c.Param("app_name")
		appVersion := c.Param("app_version")
		appBuild := c.Param("app_build")
		fileName := c.Param("file_name")

		if appName == "" || appVersion == "" || appBuild == "" || fileName == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing parameters: app name, version, build or file name"})
		}

		// find this item in db and increase download count
		app_release, err := client.AppRelease.FindFirst(
			db.AppRelease.AppName.Equals(appName),
			db.AppRelease.AppVersion.Equals(appVersion),
			db.AppRelease.AppBuild.Equals(appBuild),
		).Exec(c.Request().Context())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		_, err = client.AppRelease.FindUnique(
			db.AppRelease.ID.Equals(app_release.ID),
		).Update(
			db.AppRelease.DownloadCount.Set(app_release.DownloadCount + 1),
		).Exec(c.Request().Context())
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		filePath := fmt.Sprintf("%s/%s/%s/%s/%s", FILE_STORAGE_PATH, appName, appVersion, appBuild, fileName)
		return c.File(filePath)
	})

	e.GET("/apps/:app_name/latest", func(c echo.Context) error {
		// get the metadata for the latest release
		appName := c.Param("app_name")
		if appName == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing parameter: app name"})
		}

		release, err := client.AppRelease.FindFirst(
			db.AppRelease.AppName.Equals(appName),
		).OrderBy(
			db.AppRelease.CreatedAt.Order(db.SortOrderDesc),
		).Exec(c.Request().Context())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		if release == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No release found"})
		}

		requestHost := c.Request().Host

		latest_model := common.AppLatestModel{
			AppVersion:  release.AppVersion,
			AppBuild:    release.AppBuild,
			AppName:     release.AppName,
			DownloadURL: fmt.Sprintf("https://%s/apps/%s/%s/%s/%s", requestHost, release.AppName, release.AppVersion, release.AppBuild, release.MainFileName),
		}

		return c.JSON(http.StatusOK, latest_model)
	})

	e.Start(fmt.Sprintf(":%s", *listenPort))
}

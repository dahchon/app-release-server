package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type AppDetails struct {
	AppVersion string `json:"app_version"`
	AppBuild   string `json:"app_build"`
	AppName    string `json:"app_name"`
	GitCommit  string `json:"git_commit,omitempty"`
}

var (
	BACKEND_USERNAME = os.Getenv("ARS_BACKEND_USERNAME")
	BACKEND_PASSWORD = os.Getenv("ARS_BACKEND_PASSWORD")
)

func main() {
	// Define flags
	filePath := flag.String("file", "", "Path to the file being uploaded")
	backendURL := flag.String("url", "http://localhost:8080/admin/upload/app", "Backend URL")
	appName := flag.String("name", "", "Application name")
	appBuild := flag.String("build", "", "Application build")
	appVersion := flag.String("version", "", "Application version")
	gitCommit := flag.String("commit", "", "Git commit")

	// Parse flags
	flag.Parse()

	if *filePath == "" {
		fmt.Println("You must provide a file path")
		return
	}

	if *appName == "" || *appBuild == "" || *appVersion == "" {
		fmt.Println("You must provide the application name, build, and version")
		return
	}

	// Open the file
	file, err := os.Open(*filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Create a buffer to store our request body
	body := &bytes.Buffer{}

	// Create a multipart writer
	writer := multipart.NewWriter(body)

	// Create a form field for the file
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Copy the file into the fileWriter
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prepare the details
	details := AppDetails{
		AppName:    *appName,
		AppBuild:   *appBuild,
		AppVersion: *appVersion,
		GitCommit:  *gitCommit,
	}
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add the other fields
	_ = writer.WriteField("details", string(detailsJSON))

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		fmt.Println(err)
		return
	}

	// Create a new request
	req, err := http.NewRequest("POST", *backendURL, body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set the content type, this is very important
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add basic authentication
	req.SetBasicAuth(BACKEND_USERNAME, BACKEND_PASSWORD)

	// Do the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// Check the response
	if res.StatusCode == http.StatusOK {
		fmt.Println("File uploaded successfully")
	} else {
		fmt.Println("Failed to upload file")
		os.Exit(1)
	}
}

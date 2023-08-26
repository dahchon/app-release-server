package common

import (
	"log"
	"os"
)

var (
	BACKEND_USERNAME = os.Getenv("ARS_BACKEND_USERNAME")
	BACKEND_PASSWORD = os.Getenv("ARS_BACKEND_PASSWORD")
)

func GetBackendUsername() string {
	if BACKEND_USERNAME == "" {
		log.Fatalln("You must provide a backend username: $ARS_BACKEND_USERNAME")
		return ""
	} else {
		return BACKEND_USERNAME
	}
}

func GetBackendPassword() string {
	if BACKEND_PASSWORD == "" {
		log.Fatalln("You must provide a backend password: $ARS_BACKEND_PASSWORD")
		return ""
	} else {
		return BACKEND_PASSWORD
	}
}

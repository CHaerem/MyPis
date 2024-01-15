package utils

import (
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
)

type EnvVars struct {
	TailnetName       string
	OauthClientID     string
	OauthClientSecret string
	PiUser            string
	PiPassword        string
	WifiNetworks      []string
}

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (sh *SystemHandler) LoadEnvVars() EnvVars {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get environment variables
	tailnetName := os.Getenv("TAILNET_NAME")
	oauthClientID := os.Getenv("OAUTH_CLIENT_ID")
	oauthClientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	piUser := os.Getenv("PI_USER")
	piPassword := os.Getenv("PI_PASSWORD")
	wifiNetworks := strings.Split(os.Getenv("WIFI_NETWORKS"), " ")

	return EnvVars{
		TailnetName:       tailnetName,
		OauthClientID:     oauthClientID,
		OauthClientSecret: oauthClientSecret,
		PiUser:            piUser,
		PiPassword:        piPassword,
		WifiNetworks:      wifiNetworks,
	}
}

func (sh *SystemHandler) CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

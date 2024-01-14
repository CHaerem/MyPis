package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

type AuthKeyRequest struct {
	Capabilities struct {
		Devices struct {
			Create struct {
				Reusable      bool     `json:"reusable"`
				Ephemeral     bool     `json:"ephemeral"`
				Preauthorized bool     `json:"preauthorized"`
				Tags          []string `json:"tags"`
			} `json:"create"`
		} `json:"devices"`
	} `json:"capabilities"`
	ExpirySeconds int    `json:"expirySeconds"`
	Description   string `json:"description"`
}

func createOAuthClient(clientID, clientSecret string) *http.Client {
	log.Println("Creating OAuth client...")
	var oauthConfig = &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     "https://api.tailscale.com/api/v2/oauth/token",
	}

	return oauthConfig.Client(context.Background())
}

func createAuthKeyRequest() AuthKeyRequest {
	log.Println("Creating AuthKey request...")
	return AuthKeyRequest{
		Capabilities: struct {
			Devices struct {
				Create struct {
					Reusable      bool     `json:"reusable"`
					Ephemeral     bool     `json:"ephemeral"`
					Preauthorized bool     `json:"preauthorized"`
					Tags          []string `json:"tags"`
				} `json:"create"`
			} `json:"devices"`
		}{
			Devices: struct {
				Create struct {
					Reusable      bool     `json:"reusable"`
					Ephemeral     bool     `json:"ephemeral"`
					Preauthorized bool     `json:"preauthorized"`
					Tags          []string `json:"tags"`
				} `json:"create"`
			}{
				Create: struct {
					Reusable      bool     `json:"reusable"`
					Ephemeral     bool     `json:"ephemeral"`
					Preauthorized bool     `json:"preauthorized"`
					Tags          []string `json:"tags"`
				}{
					Reusable:      false,
					Ephemeral:     false,
					Preauthorized: true,
					Tags:          []string{"tag:MyPis"},
				},
			},
		},
		ExpirySeconds: 86400,
		Description:   "Created for MyPis",
	}
}

func sendRequest(client *http.Client, url string, jsonData []byte) *http.Response {
	log.Println("Sending request...")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error creating authkey: %v", err)
	}

	return resp
}

func GetAuthKey(clientID, clientSecret, tailnet string) string {
	client := createOAuthClient(clientID, clientSecret)

	// Create a new authkey
	url := fmt.Sprintf("https://api.tailscale.com/api/v2/tailnet/%s/keys", tailnet)

	authKeyRequest := createAuthKeyRequest()

	jsonData, err := json.Marshal(authKeyRequest)
	if err != nil {
		log.Fatalf("error marshalling JSON: %v", err)
	}

	resp := sendRequest(client, url, jsonData)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response body: %v", err)
	}

	return string(body)
}

package main

import (
	"os"
	"testing"
)

func TestLoadEnvVars(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("TAILNET_NAME", "testTailnetName")
	os.Setenv("OAUTH_CLIENT_ID", "testOauthClientID")
	os.Setenv("OAUTH_CLIENT_SECRET", "testOauthClientSecret")
	os.Setenv("PI_USER", "testPiUser")
	os.Setenv("PI_PASSWORD", "testPiPassword")
	os.Setenv("WIFI_NETWORKS", "testWifiNetwork1 testWifiNetwork2")

	// Call loadEnvVars
	env := loadEnvVars()

	// Assert the expected results
	if env.TailnetName != "testTailnetName" {
		t.Errorf("Expected testTailnetName, got %s", env.TailnetName)
	}
	if env.OauthClientID != "testOauthClientID" {
		t.Errorf("Expected testOauthClientID, got %s", env.OauthClientID)
	}
	if env.OauthClientSecret != "testOauthClientSecret" {
		t.Errorf("Expected testOauthClientSecret, got %s", env.OauthClientSecret)
	}
	if env.PiUser != "testPiUser" {
		t.Errorf("Expected testPiUser, got %s", env.PiUser)
	}
	if env.PiPassword != "testPiPassword" {
		t.Errorf("Expected testPiPassword, got %s", env.PiPassword)
	}
	if len(env.WifiNetworks) != 2 || env.WifiNetworks[0] != "testWifiNetwork1" || env.WifiNetworks[1] != "testWifiNetwork2" {
		t.Errorf("Expected [testWifiNetwork1 testWifiNetwork2], got %v", env.WifiNetworks)
	}
}

func TestCommandExists(t *testing.T) {
	// Test with a command that should exist
	if !commandExists("ls") {
		t.Errorf("Expected true, got false")
	}

	// Test with a command that should not exist
	if commandExists("nonexistentcommand") {
		t.Errorf("Expected false, got true")
	}
}

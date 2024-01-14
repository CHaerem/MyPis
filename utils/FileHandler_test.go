package utils

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestFileHandlerFileExists(t *testing.T) {
	// Create a temporary file for testing
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Create a FileHandler instance
	fh := FileHandler{}

	// Test the FileExists method
	exists := fh.FileExists(file.Name())
	if !exists {
		t.Errorf("Expected file to exist, but it doesn't")
	}
}

func TestFileHandlerDownloadFile(t *testing.T) {
	// Create a temporary file for testing
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Create a FileHandler instance
	fh := FileHandler{}

	// Start a local HTTP server to serve the test file
	server := &http.Server{
		Addr:    ":8080",
		Handler: http.FileServer(http.Dir(".")),
	}

	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	defer func() {
		if err := server.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	// Test the DownloadFile method
	err = fh.DownloadFile(file.Name(), "http://localhost:8080/testfile")
	if err != nil {
		t.Errorf("Failed to download file: %v", err)
	}

	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}
}

func TestFileHandlerExtractImageFile(t *testing.T) {
	// Create a temporary file for testing
	file, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Create a FileHandler instance
	fh := FileHandler{}

	// Test the ExtractImageFile method
	imagePath, err := fh.ExtractImageFile(file.Name())
	if err != nil {
		t.Errorf("Failed to extract image file: %v", err)
	}

	// Verify that the extracted image file exists
	exists := fh.FileExists(imagePath)
	if !exists {
		t.Errorf("Expected extracted image file to exist, but it doesn't")
	}
}

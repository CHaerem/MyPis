package utils

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDownloadFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Create a mock HTTP server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve a sample file
		http.ServeFile(w, r, "testdata/samplefile.txt")
	}))
	defer mockServer.Close()

	// Create a new instance of FileHandler
	fh := &FileHandler{}

	// Call the DownloadFile method
	err = fh.DownloadFile(tempFile.Name(), mockServer.URL)
	if err != nil {
		t.Errorf("DownloadFile returned an error: %v", err)
	}

	// Verify that the file was downloaded correctly
	fileInfo, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to get file info: %v", err)
	}
	if fileInfo.Size() != 13 {
		t.Errorf("Downloaded file size is incorrect. Expected 13, got %d", fileInfo.Size())
	}
}

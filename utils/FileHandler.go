package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type FileHandler struct{}

func (fh *FileHandler) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func (fh *FileHandler) DownloadFile(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (fh *FileHandler) ExtractImageFile(fileName string) (string, error) {
	imgFile := strings.TrimSuffix(fileName, ".xz")
	if !fh.FileExists(imgFile) {
		fmt.Println("Extracting the .img file...")
		cmd := exec.Command("xz", "-dv", fileName)
		err := cmd.Run()
		if err != nil {
			return "", fmt.Errorf("error occurred during file extraction: %v", err)
		}
	} else {
		fmt.Printf("The file %s already exists.\n", imgFile)
	}
	return imgFile, nil
}

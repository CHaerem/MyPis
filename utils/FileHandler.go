package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type FileHandler struct{}

type WriteCounter struct {
	Total     uint64
	Start     time.Time
	TotalSize uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	atomic.AddUint64(&wc.Total, uint64(n))
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	duration := time.Since(wc.Start)
	bytesPerSec := float64(wc.Total) / duration.Seconds()
	remainingSec := float64(wc.TotalSize-wc.Total) / bytesPerSec
	fmt.Printf("\r%s", strings.Repeat(" ", 50))
	fmt.Printf("\rDownloading... %v of %v bytes (%.2f%%) complete. ETA: %.2fs", wc.Total, wc.TotalSize, float64(wc.Total)*100/float64(wc.TotalSize), remainingSec)
}

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

	totalSize, err := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	counter := &WriteCounter{Start: time.Now(), TotalSize: totalSize}
	if _, err = io.Copy(out, io.TeeReader(resp.Body, counter)); err != nil {
		return err
	}

	fmt.Print("\n")
	return nil
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
		fmt.Println("Image file already exists.")
	}
	return imgFile, nil
}

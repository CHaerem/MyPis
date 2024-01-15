package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type DiskHandler struct{}

func NewDiskHandler() *DiskHandler {
	return &DiskHandler{}
}

func (dh *DiskHandler) GetDevicePath() string {
	// Get the list of all disk devices
	out, err := exec.Command("diskutil", "list").Output()
	if err != nil {
		fmt.Println("Error getting device paths:", err)
		os.Exit(1)
	}

	// Use a regular expression to extract the device identifiers
	re := regexp.MustCompile(`(/dev/disk\d+)`)
	matches := re.FindAllStringSubmatch(string(out), -1)

	devicePaths := []string{}
	for _, match := range matches {
		devicePath := match[1]

		// Get device info
		info, err := exec.Command("diskutil", "info", devicePath).Output()
		if err != nil {
			fmt.Println("Error getting device info:", err)
			os.Exit(1)
		}

		// Exclude non-removable devices
		if !strings.Contains(string(info), "Removable Media: No") {
			devicePaths = append(devicePaths, devicePath)
		}
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Choose your device:")
	for i, devicePath := range devicePaths {
		// Get device info
		info, err := exec.Command("diskutil", "info", devicePath).Output()
		if err != nil {
			fmt.Println("Error getting device info:", err)
			os.Exit(1)
		}

		// Extract the most relevant information
		re := regexp.MustCompile(`Device Identifier:.*|Device Node:.*|Device Location:.*|Removable Media:.*|Disk Size:.*`)
		matches := re.FindAllString(string(info), -1)

		fmt.Printf("%d) %s\n", i+1, devicePath)
		for _, match := range matches {
			fmt.Println(strings.TrimSpace(match))
		}
		fmt.Println()
	}
	fmt.Print("Enter the number of your choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	choiceInt, err := strconv.Atoi(choice)
	if err != nil || choiceInt <= 0 || choiceInt > len(devicePaths) {
		fmt.Println("Error: Invalid input. Please enter a number.")
		os.Exit(1)
	}
	devicePath := devicePaths[choiceInt-1]
	return devicePath
}

func (dh *DiskHandler) FlashSDCard(imgFile string, devicePath string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to flash to %s? This will erase all data on the device. (y/n): ", devicePath)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if confirm == "y" {
		fmt.Println("Flashing the SD card...")
		// Unmount the disk
		unmountCmd := exec.Command("diskutil", "unmountDisk", devicePath)
		err := unmountCmd.Run()
		if err != nil {
			fmt.Printf("Error occurred during disk unmounting: %v\n", err)
			os.Exit(1)
		}
		// Get the size of the image file
		fileInfo, err := os.Stat(imgFile)
		if err != nil {
			fmt.Printf("Error occurred while getting file size: %v\n", err)
			os.Exit(1)
		}
		fileSize := fileInfo.Size()
		// Flash the SD card
		cmd := exec.Command("sh", "-c", fmt.Sprintf("pv -petra -s %d %s | sudo dcfldd of=%s bs=4M", fileSize, imgFile, devicePath))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error occurred during SD card flashing: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("SD card flashing cancelled.")
		os.Exit(0)
	}
}

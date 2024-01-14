package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/chaerem/mypis/utils"

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

func loadEnvVars() EnvVars {
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

func main() {
	// Load environment variables
	env := loadEnvVars()

	// Check for required commands
	requiredCommands := []string{"curl", "unzip", "diskutil", "dd", "ansible-playbook"}
	for _, cmd := range requiredCommands {
		if !commandExists(cmd) {
			fmt.Printf("Error: Required command '%s' is not installed.\n", cmd)
			os.Exit(1)
		}
	}

	// Ask for the hostname
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the hostname: ")
	hostname, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading hostname: %v", err)
	}
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		log.Fatal("Hostname cannot be empty")
	}

	// Get Tailscale auth key
	authKey := utils.GetAuthKey(env.OauthClientID, env.OauthClientSecret, env.TailnetName)

	fmt.Println("Tailscale Auth Key:", authKey)

	// Download the latest image
	imageURL, fileName, err := getLatestImageURL()
	if err != nil {
		fmt.Println("Error getting the latest image URL:", err)
		os.Exit(1)
	}

	fh := &utils.FileHandler{}
	if !fh.FileExists(fileName) {
		err := fh.DownloadFile(fileName, imageURL)
		if err != nil {
			fmt.Println("Error downloading the file:", err)
			os.Exit(1)
		}
	}

	// Extract the .img file from the .xz file
	imgFile, err := fh.ExtractImageFile(fileName)
	if err != nil {
		fmt.Println("Error extracting the image file:", err)
		os.Exit(1)
	}

	// Get a list of device paths
	devicePath := getDevicePath()

	// Flash the SD card
	flashSDCard(imgFile, devicePath)

	// Setup the boot partition
	setupFirstBootScript(authKey, env.PiUser, env.PiPassword, hostname, devicePath, env.WifiNetworks)
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func getLatestImageURL() (string, string, error) {
	resp, err := http.Get("https://downloads.raspberrypi.org/raspios_lite_arm64_latest")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// The final URL after following all redirects is stored in resp.Request.URL
	url := resp.Request.URL.String()

	// Extract the filename from the URL
	_, filename := path.Split(url)

	return url, filename, nil
}

func getDevicePath() string {
	// Get the list of device paths
	out, err := exec.Command("ls", "/dev/sd*").Output()
	if err != nil {
		fmt.Println("Error getting device paths:", err)
		os.Exit(1)
	}
	devicePaths := strings.Split(string(out), "\n")

	// Rest of the code is the same
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Choose your device:")
	for i, devicePath := range devicePaths {
		fmt.Printf("%d) %s\n", i+1, devicePath)
	}
	fmt.Print("Enter the number of your choice: ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	choiceInt, err := strconv.Atoi(choice)
	if err != nil {
		fmt.Println("Error: Invalid input. Please enter a number.")
		os.Exit(1)
	}
	devicePath := devicePaths[choiceInt-1]
	return devicePath
}

func flashSDCard(imgFile string, devicePath string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Are you sure you want to flash to %s? This will erase all data on the device. (y/n): ", devicePath)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if confirm == "y" {
		fmt.Println("Flashing the SD card...")
		cmd := exec.Command("dd", "if="+imgFile, "of="+devicePath, "bs=1m")
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error occurred during SD card flashing.")
			os.Exit(1)
		}
	} else {
		fmt.Println("SD card flashing cancelled.")
		os.Exit(0)
	}
}

func setupFirstBootScript(authKey, piUser, piPassword, hostname, devicePath string, wifiNetworks []string) {
	// Load the template
	tmpl, err := ioutil.ReadFile("firstboot_template.sh")
	if err != nil {
		fmt.Println("Error reading the template file:", err)
		os.Exit(1)
	}

	data := struct {
		AuthKey      string
		PiUser       string
		PiPassword   string
		Hostname     string
		WifiNetworks []struct{ SSID, PSK string }
	}{
		AuthKey:    authKey,
		PiUser:     piUser,
		PiPassword: piPassword,
		Hostname:   hostname,
	}

	for _, wifi := range wifiNetworks {
		parts := strings.SplitN(wifi, ",", 2)
		data.WifiNetworks = append(data.WifiNetworks, struct{ SSID, PSK string }{parts[0], parts[1]})
	}

	// Generate the script
	t := template.Must(template.New("firstboot").Parse(string(tmpl)))
	var script bytes.Buffer
	if err := t.Execute(&script, data); err != nil {
		fmt.Println("Error generating the first boot script:", err)
		os.Exit(1)
	}

	// Write the script to a file
	if err := ioutil.WriteFile("firstboot.sh", script.Bytes(), 0755); err != nil {
		fmt.Println("Error writing the first boot script to a file:", err)
		os.Exit(1)
	}

	// Mount the boot partition
	bootPartition := devicePath + "1"
	if err := exec.Command("mount", bootPartition, "/mnt").Run(); err != nil {
		fmt.Printf("Error mounting the boot partition %s: %v\n", bootPartition, err)
		os.Exit(1)
	}

	// Copy the script to the boot partition
	if err := exec.Command("cp", "firstboot.sh", "/mnt").Run(); err != nil {
		fmt.Println("Error copying the first boot script to the boot partition:", err)
		os.Exit(1)
	}

	// Unmount the boot partition
	if err := exec.Command("umount", "/mnt").Run(); err != nil {
		fmt.Println("Error unmounting the boot partition:", err)
		os.Exit(1)
	}

	// Configure the Raspberry Pi to run the script on boot
	rootPartition := devicePath + "2"
	if err := exec.Command("mount", rootPartition, "/mnt").Run(); err != nil {
		fmt.Printf("Error mounting the root partition %s: %v\n", rootPartition, err)
		os.Exit(1)
	}

	if err := exec.Command("bash", "-c", `echo "@reboot root /boot/firstboot.sh" >> /mnt/etc/crontab`).Run(); err != nil {
		fmt.Println("Error configuring the Raspberry Pi to run the first boot script on boot:", err)
		os.Exit(1)
	}

	if err := exec.Command("umount", "/mnt").Run(); err != nil {
		fmt.Println("Error unmounting the root partition:", err)
		os.Exit(1)
	}
}

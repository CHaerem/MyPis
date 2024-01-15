package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/chaerem/mypis/utils"
)

func main() {
	// Create a SystemHandler
	sh := utils.NewSystemHandler()

	// Load environment variables
	env := sh.LoadEnvVars()

	// Check for required commands
	requiredCommands := []string{"curl", "unzip", "diskutil", "dd", "ansible-playbook"}
	for _, cmd := range requiredCommands {
		if !sh.CommandExists(cmd) {
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

	// Create a DiskHandler
	dh := utils.NewDiskHandler()

	// Get a list of device paths
	devicePath := dh.GetDevicePath()

	// Flash the SD card
	dh.FlashSDCard(imgFile, devicePath)

	// Setup the boot partition
	// setupFirstBootScript(authKey, env.PiUser, env.PiPassword, hostname, devicePath, env.WifiNetworks)
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

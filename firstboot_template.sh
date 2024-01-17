#!/bin/bash
sudo tailscale up --authkey={{.AuthKey}}

# Enable SSH
sudo systemctl enable ssh
sudo systemctl start ssh

# Change default user password
echo "{{.PiUser}}:{{.PiPassword}}" | sudo chpasswd

# Set up WiFi
sudo bash -c 'cat >> /etc/wpa_supplicant/wpa_supplicant.conf' << EOF
{{range .WifiNetworks}}
network={
    ssid="{{.SSID}}"
    psk="{{.PSK}}"
}
{{end}}
EOF

# Reconfigure wireless interface
wpa_cli -i wlan0 reconfigure

# Set hostname
sudo hostname {{.Hostname}}
echo "{{.Hostname}}" | sudo tee /etc/hostname

# Update and upgrade the system
sudo apt-get update
sudo apt-get upgrade -y
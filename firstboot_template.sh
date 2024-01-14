#!/bin/bash
sudo tailscale up --authkey={{AUTH_KEY}}

# Enable SSH
sudo systemctl enable ssh
sudo systemctl start ssh

# Change default user password
echo "{{PI_USER}}:{{PI_PASSWORD}}" | sudo chpasswd

# Set up WiFi
sudo bash -c 'cat >> /etc/wpa_supplicant/wpa_supplicant.conf' << EOF
$(for wifi in {{WIFI_NETWORKS}}; do
ssid=$(echo $wifi | cut -d',' -f1)
psk=$(echo $wifi | cut -d',' -f2)
cat << WIFI
network={
    ssid="${ssid}"
    psk="${psk}"
}
WIFI
done)
EOF

# Reconfigure wireless interface
wpa_cli -i wlan0 reconfigure

# Set hostname
sudo hostname {{HOSTNAME}}
echo "{{HOSTNAME}}" | sudo tee /etc/hostname

# Update and upgrade the system
sudo apt-get update
sudo apt-get upgrade -y

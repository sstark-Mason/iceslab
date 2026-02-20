#!/bin/bash


# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

sudo -u $SUDO_USER pactl set-sink-volume @DEFAULT_SINK@ 0

# Set max parallel downloads for DNF
tee /etc/dnf/dnf.conf > /dev/null <<EOF
[main]
fastestmirror=True
max_parallel_downloads=10
EOF

# Update software repositories
dnf install -y https://mirrors.rpmfusion.org/free/fedora/rpmfusion-free-release-$(rpm -E %fedora).noarch.rpm
dnf install -y https://mirrors.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-$(rpm -E %fedora).noarch.rpm

# Set flatpak remote repositories
flatpak remote-delete fedora
flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo

# Install basic software
dnf install -y fish curl rsync git firefox chromium code

# Updates
dnf group upgrade core -y
dnf check-update
dnf upgrade -y

# Update firmware
fwupdmgr get-devices
fwupdmgr refresh --force
fwupdmgr get-updates
fwupdmgr update -y

# Basic drivers and Vulkan support
dnf install -y mesa-dri-drivers mesa-vulkan-drivers vulkan-loader mesa-libGLU
dnf install -y intel-media-driver

# Swap to full ffmpeg
dnf swap -y ffmpeg-free ffmpeg --allowerasing

# Install all the GStreamer plugins
dnf install -y gstreamer1-plugins-{bad-\*,good-\*,base} \
  gstreamer1-plugin-openh264 gstreamer1-libav lame\* \
  --exclude=gstreamer1-plugins-bad-free-devel

# Install multimedia groups
dnf group install -y multimedia
dnf group install -y sound-and-video

# Install VA-API stuff
dnf install -y ffmpeg-libs libva libva-utils

# Install the Cisco codec (it's free but weird licensing)
dnf install -y openh264 gstreamer1-plugin-openh264 mozilla-openh264

# Install Microsoft fonts
dnf install -y cabextract xorg-x11-font-utils fontconfig
rpm -i --nodigest --nosignature https://downloads.sourceforge.net/project/mscorefonts2/rpms/msttcore-fonts-installer-2.6-1.noarch.rpm
fc-cache -fv

#!/bin/bash

# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Read station number from /etc/iceslab/iceslab.conf
STATION_NUM=$(grep "station_number=" /etc/iceslab/iceslab.conf | cut -d'=' -f2)
if test -z "$STATION_NUM"; then
    echo "Station number not found in /etc/iceslab/iceslab.conf. Please enter (e.g., 01): "
    read STATION_NUM
    STATION_NUM=$(echo "$STATION_NUM" | xargs) # Trim whitespace
    # Update /etc/iceslab/iceslab.conf with station number
    sudo tee -a /etc/iceslab/iceslab.conf > /dev/null <<EOF
        [ID]
        station_number=$STATION_NUM
        EOF
fi

sudo hostnamectl set-hostname "iceslab-$STATION_NUM"

sudo -i -u $SUDO_USER <<EOF

    # Use Breeze Dark (not sure which to use)
    kwriteconfig6 --file kdeglobals --group Icons --key Theme breeze-dark
    kwriteconfig6 --file kdeglobals --group General --key ColorScheme "BreezeDark"

    # Disable KDE login splash
    touch ~/.config/ksplashrc
    kwriteconfig6 --file ksplashrc --group KSplash --key Theme None

    # Edit logout hotkey
    touch ~/.config/kglobalshortcutsrc
    kwriteconfig6 --file kglobalshortcutsrc --group kwin --key "LogOut" "Meta+Shift+L,none,Log Out"

    # Disable mouse acceleration
    touch ~/.config/kcminputrc
    kwriteconfig6 --file kcminputrc --group Mouse --key PointerAcceleration "\-0.200"
    kwriteconfig6 --file kcminputrc --group Mouse --key PointerAccelerationProfile 1

    kwriteconfig6 --file kcminputrc --group Libinput 16700 9492 Dell Computer Corp Dell Universal Receiver Mouse --key PointerAcceleration "\-0.200"
    kwriteconfig6 --file kcminputrc --group Libinput 16700 9492 Dell Computer Corp Dell Universal Receiver Mouse --key PointerAccelerationProfile 1

EOF

# Copy configs to etc/skel
ORIGINAL_HOME=$(eval echo "~$SUDO_USER")

mkdir -p /etc/skel/.config
cp "$ORIGINAL_HOME/.config/kdeglobals" /etc/skel/.config/
chown root:root /etc/skel/.config/kdeglobals

GUEST_USER="guest-$STATION_NUM"

#!/bin/bash
GUEST_USER="guest"
GUEST_HOME="/home/$GUEST_USER/"
TEMPLATE_DIR="/opt/iceslab/guest-template/"

# Check if the script is run as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Terminate and delete guest user if it exists
loginctl terminate-user "$GUEST_USER" 2>/dev/null
userdel --remove "$GUEST_USER" 2>/dev/null

# Create guest user
useradd -m "$GUEST_USER"
passwd -d "$GUEST_USER"

# rsync guest home to template
rsync --delete --recursive  $GUEST_HOME $TEMPLATE_DIR

# Place guest session manager service
rsync /opt/iceslab/assets/services/guest-session-management.service /etc/systemd/system/

# Allow guest user to run the iceslab script without password
cat <<EOF > /etc/sudoers.d/iceslab
$GUEST_USER ALL=(root) NOPASSWD: /opt/iceslab/iceslab
EOF
chmod 0440 /etc/sudoers.d/iceslab

systemctl daemon-reload
systemctl enable guest-session-management.service

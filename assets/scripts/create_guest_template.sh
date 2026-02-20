#!/bin/bash
GUEST_USER="guest"
GUEST_HOME="/home/$GUEST_USER/"
TEMPLATE_DIR="/opt/iceslab/guest-template/"

# Terminate and delete guest user if it exists
loginctl terminate-user "$GUEST_USER" 2>/dev/null
userdel --remove "$GUEST_USER" 2>/dev/null

# Create guest user
useradd -m "$GUEST_USER"
passwd -d "$GUEST_USER"

# rsync guest home to template
rsync --delete  $GUEST_HOME $TEMPLATE_DIR

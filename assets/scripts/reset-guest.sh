#!/bin/bash
GUEST_USER="guest"
GUEST_HOME="/home/$GUEST_USER/"
TEMPLATE_DIR="/opt/guest-template/"

systemctl terminate-user "$GUEST_USER" 2>/dev/null

# 1. Create guest user if it doesn't exist
if ! id "$GUEST_USER" &>/dev/null; then
    useradd -m "$GUEST_USER"
    passwd -d "$GUEST_USER" # Clear password for easy login
fi

# 2. Wipe current guest home and sync from template with rsync
rsync -a --delete  $TEMPLATE_DIR $GUEST_HOME

# 3. Fix ownership
chown -R $GUEST_USER:$GUEST_USER $GUEST_HOME
restorecon -R $GUEST_HOME

#!/bin/bash
GUEST_USER="guest"
TEMPLATE_USER="guest-template"

# 1. Create guest user if it doesn't exist
if ! id "$GUEST_USER" &>/dev/null; then
    useradd -m "$GUEST_USER"
    passwd -d "$GUEST_USER" # Clear password for easy login
fi

# 2. Wipe current guest home and sync from template
rm -rf /home/$GUEST_USER/*
rm -rf /home/$GUEST_USER/.* 2>/dev/null
cp -rT /home/$TEMPLATE_USER/ /home/$GUEST_USER/

# 3. Fix ownership
chown -R $GUEST_USER:$GUEST_USER /home/$GUEST_USER
restorecon -R /home/$GUEST_USER

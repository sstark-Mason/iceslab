#!/bin/bash
GUEST_USER="guest"
GUEST_HOME="/home/$GUEST_USER/"
TEMPLATE_DIR="/opt/iceslab/guest-template/"

# Create guest user if it doesn't exist
# if ! id "$GUEST_USER"; then
#     useradd -m "$GUEST_USER"
#     passwd -d "$GUEST_USER"
# fi

# Wipe current guest home and sync from template with rsync
rsync --delete  $TEMPLATE_DIR $GUEST_HOME

# Fix ownership
chown -R $GUEST_USER:$GUEST_USER $GUEST_HOME
restorecon -R $GUEST_HOME

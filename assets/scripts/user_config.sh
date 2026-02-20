
touch ~/.config/kdeglobals
kwriteconfig6 --file kdeglobals --group General --key ColorScheme "Breeze Dark" # ???
kwriteconfig6 --file kdeglobals --group Icons --key Theme "Breeze Dark" # ???

touch .config/ksplashrc
kwriteconfig6 --file ksplashrc --group KSplash --key Theme None # Yes?

# touch .config/kglobalshortcutsrc
# kwriteconfig6 --file kglobalshortcutsrc --group kwin --key "LogOut" "Meta+Shift+L,,Log Out Without Confirmation"

touch ~/.config/kcminputrc
kwriteconfig6 --file kcminputrc --group Mouse --key PointerAcceleration -- -0.200
kwriteconfig6 --file kcminputrc --group Mouse --key PointerAccelerationProfile 1
kwriteconfig6 --file kcminputrc --group Libinput 16700 9492 Dell Computer Corp Dell Universal Receiver Mouse --key PointerAcceleration -- -0.200
kwriteconfig6 --file kcminputrc --group Libinput 16700 9492 Dell Computer Corp Dell Universal Receiver Mouse --key PointerAccelerationProfile 1

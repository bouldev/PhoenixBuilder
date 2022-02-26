ICON="-icon unbundled_assets/Icon.png"
VERSION="-app-version 0.0.4"
APPID=' -app-id fastbuilder.fyne.gui'
APPBUILD='-app-build 197'
ARGS="$ICON $VERSION $APPID $APPBUILD"

fyne-cross linux -arch=amd64 $ARGS -name "fastbuilder_fyne_gui"
fyne-cross darwin -arch=amd64 $ARGS  -name "fastbuilder_fyne_gui"
fyne-cross windows -arch=amd64 $ARGS -name "fastbuilder_fyne_gui.exe"
fyne-cross android -arch=arm64 $ARGS -name "fastbuilder_fyne_gui"
fyne-cross ios $ARGS -name "fastbuilder-fyne-gui"
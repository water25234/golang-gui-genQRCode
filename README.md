# golang-gui-genQRCode


## Run Project
 - make localBuild
 - make buildMacos

## make icon
mkdir tmp.iconset

sips -z 16 16     go.png --out tmp.iconset/icon_16x16.png
sips -z 32 32     go.png --out tmp.iconset/[email protected]
sips -z 32 32     go.png --out tmp.iconset/icon_32x32.png
sips -z 64 64     go.png --out tmp.iconset/[email protected]
sips -z 128 128   go.png --out tmp.iconset/icon_128x128.png
sips -z 256 256   go.png --out tmp.iconset/[email protected]
sips -z 256 256   go.png --out tmp.iconset/icon_256x256.png
sips -z 512 512   go.png --out tmp.iconset/[email protected]
sips -z 512 512   go.png --out tmp.iconset/icon_512x512.png
sips -z 1024 1024   go.png --out tmp.iconset/[email protected]

iconutil -c icns tmp.iconset -o Icon.icns
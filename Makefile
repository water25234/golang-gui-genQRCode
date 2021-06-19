localBuild: 
	go run gen.go
	go build -o gui

buildMacos:
	./build-macos.sh

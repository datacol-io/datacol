
BUILD_CMD = go build -i cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go cmd/helper.go cmd/run.go

all: osx linux sync

osx: dist
	env GOOS=darwin GOARCH=386 $(BUILD_CMD)
	zip dist/osx.zip main

linux: dist
	env GOOS=linux GOARCH=amd64 $(BUILD_CMD)
	zip dist/linux.zip main

dist:
	mkdir -p dist

sync:
	gsutil cp -a public-read dist/* gs://datacol-distros
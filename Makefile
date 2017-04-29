
BUILD_CMD = go build -i cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go cmd/helper.go cmd/run.go cmd/infra.go cmd/upgrade.go

all: osx linux sync

osx: dist
	env GOOS=darwin GOARCH=386 $(BUILD_CMD)
	mv main datacol
	zip dist/osx.zip datacol
	rm datacol

linux: dist
	env GOOS=linux GOARCH=amd64 $(BUILD_CMD)
	mv main datacol
	zip dist/linux.zip datacol
	rm datacol

dist:
	mkdir -p dist

sync:
	gsutil cp -a public-read dist/* gs://datacol-distros

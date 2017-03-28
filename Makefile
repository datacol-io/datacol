
BUILD_CMD = go build -i cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go cmd/helper.go

osx: dist
	env GOOS=darwin GOARCH=386 $(BUILD_CMD)
	mv main dist/osx

linux: dist
	env GOOS=linux GOARCH=arm $(BUILD_CMD)
	mv main dist/linux

dist:
	mkdir -p dist

sync:
	gsutil cp -a public-read dist/* gs://datacol-distros
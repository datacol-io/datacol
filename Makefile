
BUILD_CMD=cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go cmd/helper.go cmd/run.go cmd/infra.go cmd/upgrade.go cmd/login.go
version=1.0.0-alpha.4
MODEL_PROTO_DIR=./api/models
SERVICE_PROTO_DIR=./api/controller
VEDNOR_GOOGLE_APIS=./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

zip:
	env VERSION=${version} ruby hack/make.rb push_zip

all:
	env VERSION=${version} ruby hack/make.rb

vet:
	go vet cmd/*.go
	go vet cmd/provider/gcp/*.go
	go vet client/*.go
	go vet cloud/google/*.go
	go vet cloud/*.go
	go vet api/*.go
	go vet api/models/*.go
	go vet api/controller/*.go

	go fmt cmd/*.go
	go fmt cmd/provider/gcp/*.go
	go fmt client/*.go
	go fmt cloud/google/*.go
	go fmt cloud/*.go
	go fmt api/*.go
	go fmt api/models/*.go
	go fmt api/controller/*.go

build:
	go build -ldflags="-s -w" -i ${BUILD_CMD}

apictl:
	go build -o apictl -ldflags="-s -w" -i api/*.go

gen:
	## building api/models/*.proto
	protoc -I $(GOPATH)/src -I ./vendor/ \
		-I $(MODEL_PROTO_DIR) \
		-I $(VEDNOR_GOOGLE_APIS) \
		--go_out=plugins=grpc:$(MODEL_PROTO_DIR) \
		$(MODEL_PROTO_DIR)/*.proto

  #building api/controller/*.proto
	protoc -I $(GOPATH)/src -I ./vendor/ \
		-I $(SERVICE_PROTO_DIR) \
		-I $(VEDNOR_GOOGLE_APIS) \
		--go_out=plugins=grpc:$(SERVICE_PROTO_DIR) \
		$(SERVICE_PROTO_DIR)/*.proto

	protoc -I $(GOPATH)/src -I ./vendor/ \
		-I $(SERVICE_PROTO_DIR) \
		-I $(VEDNOR_GOOGLE_APIS) \
    --grpc-gateway_out=logtostderr=true:$(SERVICE_PROTO_DIR) \
		$(SERVICE_PROTO_DIR)/*.proto

	protoc \
		-I $(GOPATH)/src -I ./vendor/ \
 		-I $(SERVICE_PROTO_DIR) \
		-I $(VEDNOR_GOOGLE_APIS) \
    --swagger_out=logtostderr=true:$(SERVICE_PROTO_DIR) \
    $(SERVICE_PROTO_DIR)/*.proto


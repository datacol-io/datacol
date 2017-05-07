
BUILD_CMD=cmd/main.go cmd/build.go cmd/stack.go cmd/apps.go cmd/deploy.go cmd/kubectl.go cmd/env.go cmd/logs.go cmd/helper.go cmd/run.go cmd/infra.go cmd/upgrade.go
version=1.0.0-alpha.3
PROTO_DIR=./api/models

zip:
	env VERSION=${version} ruby hack/make.rb push_zip

all:
	env VERSION=${version} ruby hack/make.rb

vet:
	go vet cmd/*.go
	go vet client/*.go
	go vet cloud/google/*.go
	go fmt cmd/*.go
	go fmt client/*.go
	go fmt cloud/google/*.go

build:
	go build -i ${BUILD_CMD}
	
gen:
	protoc -I $(PROTO_DIR) -I ./vendor/ \
		-I ./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:$(PROTO_DIR)/ \
		$(PROTO_DIR)/*.proto

	protoc -I $(PROTO_DIR) -I ./vendor/ \
		-I ./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --grpc-gateway_out=logtostderr=true:$(PROTO_DIR)/ \
		$(PROTO_DIR)/*.proto

	protoc \
    -I $(PROTO_DIR) -I ./vendor/ \
    -I ./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --swagger_out=logtostderr=true:$(PROTO_DIR)/ \
    $(PROTO_DIR)/*.proto


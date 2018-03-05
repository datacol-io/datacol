version=1.0.0-alpha.10
MODEL_PROTO_DIR=./api/models
SERVICE_PROTO_DIR=./api/controller
VEDNOR_GOOGLE_APIS=./vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
VEDNOR_GOLANG_PROTOS=./vendor/github.com/golang/protobuf
VEDNOR_GOGO_PROTOS=./vendor/github.com/gogo/protobuf/protobuf
PROTOC_INCLUDE_DIR=-I $(VEDNOR_GOOGLE_APIS) -I $(VEDNOR_GOLANG_PROTOS) -I $(VEDNOR_GOGO_PROTOS) -I ./vendor -I $(GOPATH)/src

zip:
	env VERSION=${version} ruby hack/make.rb push_zip

all:
	env VERSION=${version} ruby hack/make.rb

vet:
	go fmt ./cmd ./cmd/stdcli ./cmd/provider/gcp ./cmd/provider/aws ./client/ ./cloud/google/ ./cloud/aws/ ./cloud/ ./api ./api/models/ ./api/controller/
	go vet ./cmd ./cmd/provider/gcp ./client/ ./cloud/google/ ./cloud/aws/ ./cloud/ ./api ./api/models/ ./api/controller/
	goimports ./cmd ./cmd/provider/gcp ./client/ ./cloud/google/ ./cloud/aws/ ./cloud/ ./api ./api/models/ ./api/controller/

cmd:
	go build -o datacol -ldflags="-s -w" .

api:
	go build -o apictl -ldflags="-s -w" -i api/*.go

proto:
	go install -v ./vendor/github.com/golang/protobuf/protoc-gen-go
	go install -v ./vendor/github.com/gogo/protobuf/protoc-gen-gogo	
	go install -v ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	go install -v ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	go install -v ./vendor/github.com/go-swagger/go-swagger/cmd/swagger

gentest:
	protoc -I $(SERVICE_PROTO_DIR) $(PROTOC_INCLUDE_DIR) \
		--go_out=plugins=grpc:$(SERVICE_PROTO_DIR) \
		$(SERVICE_PROTO_DIR)/*.proto

gen:
	go-bindata -o cmd/provider/aws/templates.go cmd/provider/aws/templates/ && sed -i 's/main/aws/g' cmd/provider/aws/templates.go
	go-bindata -o cloud/aws/templates.go cloud/aws/templates/ && sed -i 's/main/aws/g' cloud/aws/templates.go
	go-bindata -o cloud/google/templates.go cloud/google/templates/ && sed -i 's/main/google/g' cloud/google/templates.go

	## building api/models/*.proto
	protoc -I $(MODEL_PROTO_DIR) $(PROTOC_INCLUDE_DIR) \
		--gogo_out=plugins=grpc:$(MODEL_PROTO_DIR) \
		$(MODEL_PROTO_DIR)/*.proto

  #building api/controller/*.proto
	protoc -I $(SERVICE_PROTO_DIR) \
		$(PROTOC_INCLUDE_DIR) \
		--go_out=plugins=grpc:$(SERVICE_PROTO_DIR) \
		$(SERVICE_PROTO_DIR)/*.proto

	protoc -I $(SERVICE_PROTO_DIR) \
		$(PROTOC_INCLUDE_DIR) \
    --grpc-gateway_out=logtostderr=true:$(SERVICE_PROTO_DIR) \
		$(SERVICE_PROTO_DIR)/*.proto

	protoc -I $(SERVICE_PROTO_DIR) \
		$(PROTOC_INCLUDE_DIR) \
    	--swagger_out=logtostderr=true:$(SERVICE_PROTO_DIR) \
    	$(SERVICE_PROTO_DIR)/*.proto


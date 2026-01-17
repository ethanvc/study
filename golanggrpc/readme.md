# 安装 protoc-gen-go（生成普通 Go 代码）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc（生成 gRPC 相关代码，如果需要）
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"


protoc --go_out=. -I . --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative golanggrpc.proto

test:
	go test -vv internal/...

PROTOC-exists:
	@which protoc > /dev/null

proto: PROTOC-exists
	protoc -I ./rpc \
       --go_out ./rpc/ --go_opt paths=source_relative \
       --go-grpc_out ./rpc/ --go-grpc_opt paths=source_relative \
       ./rpc/zkp/zkp.proto

_client:
	go run client/main.go

server:
	go run cmd/zkp/main.go
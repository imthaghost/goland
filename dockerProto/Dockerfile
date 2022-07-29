FROM golang:1.18

# Enviornment variables
ENV APP_NAME dockerProto

# Working directory
WORKDIR tmp

RUN apt-get update
RUN apt install -y protobuf-compiler

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

COPY . .

RUN protoc -I ./rpc \
       --go_out ./rpc/ --go_opt paths=source_relative \
       --go-grpc_out ./rpc/ --go-grpc_opt paths=source_relative \
       --python_out=./rpc/ \
       ./rpc/dockerProto/dockerProto.proto

RUN ls /tmp

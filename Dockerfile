FROM golang as builder

# git

RUN apt-get update && \
    apt-get install unzip git

# grpc support
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" && \
    curl -LO $PB_REL/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip && \
    unzip protoc-3.11.4-linux-x86_64.zip -d /usr/local && \
    export PATH=$PATH:/usr/local/bin

FROM builder as mygolangproject

WORKDIR $GOPATH/src/mygolangproject
COPY . $GOPATH/src/mygolangproject

FROM mygolangproject as grpc
RUN export GO111MODULE=on && \
    export GOPROXY=https://mirrors.aliyun.com/goproxy/ && \
    go build grpcserver/my_grpc_server.go

EXPOSE 50051
ENTRYPOINT ["./my_grpc_server"]

FROM mygolangproject as server
RUN export GO111MODULE=on && \
    export GOPROXY=https://mirrors.aliyun.com/goproxy/ && \
    go build my_http_server.go

EXPOSE 8088
ENTRYPOINT ["./my_http_server"]


FROM mygolangproject as client
RUN export GO111MODULE=on && \
    export GOPROXY=https://mirrors.aliyun.com/goproxy/ && \
    go build my_http_client.go

ENTRYPOINT ["./my_http_client"]


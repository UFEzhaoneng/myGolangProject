FROM golang

# git

RUN apt-get update && \
    apt-get install unzip git

# grpc support
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" && \
    curl -LO $PB_REL/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip && \
    unzip protoc-3.11.4-linux-x86_64.zip -d /usr/local && \
    export PATH=$PATH:/usr/local/bin

RUN export GO111MODULE=on && \
    export GOPROXY=https://mirrors.aliyun.com/goproxy/ && \
    go get google.golang.org/grpc@v1.28.1 && \
    go get github.com/golang/protobuf/protoc-gen-go && \
    export PATH=$PATH:$GOPATH/bin

WORKDIR $GOPATH/src/mygolangproject
COPY . $GOPATH/src/mygolangproject
RUN cd  $GOPATH/src/mygolangproject/proto && \
    protoc --go_out=plugins=grpc:. service.proto && \
    go build $GOPATH/src/mygolangproject/grpcserver/my_grpc_server.go

EXPOSE 8000
ENTRYPOINT ["./mygolangproject"]
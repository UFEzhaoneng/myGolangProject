FROM golang

# git

RUN apt-get update && \

    apt-get -y install git unzip build-essential autoconf libtool && \

    git config --global http.postBuffer 1048576000

# grpc support
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.10.0/protoc-3.10.0-linux-x86_64.zip
RUN unzip protoc-3.10.0-linux-x86_64.zip

RUN export GOPROXY=https://goproxy.cn && \
    export GO111MODULE=on && \
    export GOPATH=/ && \
    go get google.golang.org/grpc@v1.28.1 && \
    go get -u github.com/golang/protobuf/protoc-gen-go && \
    export PATH=$PATH:$GOPATH/bin

WORKDIR $GOPATH/src/mygolangproject
COPY . $GOPATH/src/mygolangproject
RUN go build $GOPATH/src/mygolangproject/my_http_client.go

EXPOSE 8000
ENTRYPOINT ["./mygolangproject"]
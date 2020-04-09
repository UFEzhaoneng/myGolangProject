package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"

	pb "./helloWorldProto"
)

const (
	gprcAddress = "localhost:50051"
)

func connectWithGRPC(name string) string {
	// Set up a connection to the server.
	conn, err := grpc.Dial(gprcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: string(name)})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
	return r.GetMessage()
}

func main() {

	// Hello world, the web server

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, connectWithGRPC(req.RemoteAddr))
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8088", nil))
}

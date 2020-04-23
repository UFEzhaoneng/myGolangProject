package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc"
	pb "mygolangproject/proto"
)

const (
	gprcAddress = "localhost:50051"
)

func helloHandlerFunc(name string) string {
	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: string(name)})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
	return r.GetMessage()
}

func connectWithGrpc() (*grpc.ClientConn, context.Context, context.CancelFunc) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(gprcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	return conn, ctx, cancel
}

// Hello world, the web server
func helloHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, helloHandlerFunc(req.RemoteAddr))
}

func registerInfoCheck(w http.ResponseWriter, req *http.Request) (bool, string, int, string) {
	isOk := true
	name := req.PostFormValue("name")
	if name == "" {
		io.WriteString(w, "error")
		log.Print("name error")
		isOk = false
	}
	age, err := strconv.Atoi(req.PostFormValue("age"))
	if err != nil {
		io.WriteString(w, "error")
		log.Print("age error")
		isOk = false
	}
	profession := req.PostFormValue("profession")
	if profession == "" {
		io.WriteString(w, "error")
		log.Print("profession error")
		isOk = false
	}
	return isOk, name, age, profession
}

func registerHandler(w http.ResponseWriter, req *http.Request) {
	isOk, name, age, profession := registerInfoCheck(w, req)
	if !isOk {
		io.WriteString(w, "register error")
		return
	}

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.Register(ctx, &pb.RegisterRequest{Name: name, Age: int32(age), Profession: profession})
	if err != nil {
		log.Fatalf("could not register: %v", err)
		return
	}
	log.Printf("register: %v success", r.Id)
	io.WriteString(w, strconv.Itoa(int(r.Id)))
}

func idCheck(w http.ResponseWriter, req *http.Request) (int, bool) {
	id, err := strconv.Atoi(req.PostFormValue("id"))
	if err != nil {
		log.Printf("id error: %v", err)
		io.WriteString(w, "query error")
		return 0, false
	}
	return id, true
}

func responseStudentInfo(w http.ResponseWriter, studentInfo *pb.StudentInfo) {
	io.WriteString(w,
		"id: "+strconv.Itoa(int(studentInfo.Id))+
			" name: "+studentInfo.Name+
			" age: "+strconv.Itoa(int(studentInfo.Age))+
			" profession:"+studentInfo.Profession)
}

func queryHandler(w http.ResponseWriter, req *http.Request) {
	id, res := idCheck(w, req)
	if !res {
		return
	}

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.Query(ctx, &pb.StudentId{Id: int32(id)})
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("query: %v success", id)
	responseStudentInfo(w, r)
}

func alterProfessionHandler(w http.ResponseWriter, req *http.Request) {
	id, res := idCheck(w, req)
	if !res {
		return
	}

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.AlterProfession(ctx, &pb.StudentId{Id: int32(id)})
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("alterProfession: %v success", id)
	io.WriteString(w, strconv.FormatBool(r.Res))
}

func deleteHandler(w http.ResponseWriter, req *http.Request) {
	id, res := idCheck(w, req)
	if !res {
		return
	}

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.Delete(ctx, &pb.StudentId{Id: int32(id)})
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Printf("delete student: %v success", id)
	io.WriteString(w, strconv.FormatBool(r.Res))
}

func queryListHandler(w http.ResponseWriter, _ *http.Request) {

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.QueryList(ctx, &pb.QueryRequest{})
	if err != nil {
		log.Printf("%v", err)
		return
	}
	log.Print("query list success")
	for _, studentInfo := range r.StudentInfo {
		responseStudentInfo(w, studentInfo)
	}
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/alterProfession", alterProfessionHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/queryList", queryListHandler)
	log.Fatal(http.ListenAndServe(":8088", nil))
}

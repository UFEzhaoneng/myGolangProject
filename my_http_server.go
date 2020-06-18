package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"log"

	consulapi "github.com/hashicorp/consul/api"
	pb "mygolangproject/proto"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	computerScienceAndTechnology = "计算机科学与技术"
	softwareEngineering          = "软件工程"
)

var grpcAddress string

var nameCheck = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func register() {
	config := consulapi.DefaultConfig()
	config.Address = "127.0.0.1:8500"
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("consul client error : ", err)
	}

	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "httpServerNode_1"       // 服务节点的名称
	registration.Name = "httpServer"           // 服务名称
	registration.Port = 8089                   // 服务端口
	registration.Tags = []string{"httpServer"} // tag，可以为空
	registration.Address = "127.0.0.1"         // 服务 IP

	registration.Check = &consulapi.AgentServiceCheck{
		HTTP:                           fmt.Sprintf("http://%s:%d%s", registration.Address, registration.Port, "/check"),
		Timeout:                        "3s",
		Interval:                       "5s",  // 健康检查间隔
		DeregisterCriticalServiceAfter: "30s", //check失败后30秒删除本服务，注销时间，相当于过期时间
	}
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal("register server error : ", err)
	}
}

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

func grpcDiscovery() {
	var lastIndex uint64
	config := api.DefaultConfig()
	config.Address = "127.0.0.1:8500" //consul server

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal("api new client is failed, err:", err)
		return
	}
	services, metainfo, err := client.Health().Service("grpcServer", "grpc", true, &api.QueryOptions{
		WaitIndex: lastIndex, // 同步点，这个调用将一直阻塞，直到有新的更新
	})
	if err != nil {
		logrus.Warn("error retrieving instances from Consul: %v", err)
	}
	lastIndex = metainfo.LastIndex

	//addrs := map[string]struct{}{}
	for _, service := range services {
		log.Println("service.Service.Address:", service.Service.Address, "service.Service.Port:", service.Service.Port)
		grpcAddress = service.Service.Address + ":" + strconv.Itoa(service.Service.Port)
		//addrs[net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))] = struct{}{}
	}
}

func connectWithGrpc() (*grpc.ClientConn, context.Context, context.CancelFunc) {

	// Set up a connection to the server.
	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure(), grpc.WithBlock())
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
	if name == "" || !nameCheck(name) {
		io.WriteString(w, "name error")
		log.Print("name error")
		isOk = false
	}
	age, err := strconv.Atoi(req.PostFormValue("age"))
	if err != nil || age < 10 || age > 100 {
		io.WriteString(w, "age error")
		log.Print("age error")
		isOk = false
	}
	profession, res := professionCheck(w, req)
	if !res {
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
	io.WriteString(w, "connect success")
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
	io.WriteString(w, r.Id)
}

func idCheck(w http.ResponseWriter, req *http.Request) (string, bool) {
	id := req.PostFormValue("id")
	if id == "" {
		log.Printf("id is nil")
		io.WriteString(w, "query error")
		return "", false
	}
	return id, true
}

func professionCheck(w http.ResponseWriter, req *http.Request) (string, bool) {
	profession := req.PostFormValue("profession")
	if profession == "" || (profession != computerScienceAndTechnology && profession != softwareEngineering) {
		log.Print("profession error")
		io.WriteString(w, "profession error")
		return "", false
	}
	return profession, true
}

func responseStudentInfo(w http.ResponseWriter, studentInfo *pb.StudentInfo) {
	io.WriteString(w,
		"id: "+studentInfo.Id+
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

	r, err := c.Query(ctx, &pb.StudentInfo{Id: id})
	if err != nil {
		log.Printf("%v", err)
		io.WriteString(w, "query error")
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
	profession, res := professionCheck(w, req)
	if !res {
		return
	}

	conn, ctx, cancel := connectWithGrpc()
	c := pb.NewServiceClient(conn)
	defer conn.Close()
	defer cancel()

	r, err := c.AlterProfession(ctx, &pb.StudentInfo{Id: id, Profession: profession})
	if err != nil {
		log.Printf("%v", err)
		io.WriteString(w, "alter error")
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

	r, err := c.Delete(ctx, &pb.StudentInfo{Id: id})
	if err != nil {
		log.Printf("%v", err)
		io.WriteString(w, "delete error")
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

func consulCheck(w http.ResponseWriter, req *http.Request) {

	s := "consulCheck" + "remote:" + req.RemoteAddr + " " + req.URL.String()
	fmt.Println(s)
	fmt.Fprintln(w, s)
}

func main() {
	register()
	grpcDiscovery()
	http.HandleFunc("/check", consulCheck)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/alterProfession", alterProfessionHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/queryList", queryListHandler)
	log.Fatal(http.ListenAndServe(":8089", nil))
}

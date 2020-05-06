package main

import (
	pb "../proto"
	"context"
	"errors"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

const port = ":50051"

type Server struct {
	pb.UnimplementedServiceServer
}

type student struct {
	id         string //唯一
	name       string //仅支持英文，非空
	age        int32  //非空，范围【10，100】
	profession string //枚举：计算机科学与技术/软件工程
}
type safeStudentInfo struct {
	studentInfo map[string]student
	mux         sync.RWMutex
}

var allStudentInfo = safeStudentInfo{studentInfo: make(map[string]student)}

func getUUID() string {

	v4, err := uuid.NewV4()
	if err != nil {
		log.Fatal("v4 create err")
	}
	return v4.String()
}

// SayHello implements helloworld.GreeterServer
func (s *Server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

//Register implements helloworld.GreeterServer
func (s *Server) Register(_ context.Context, info *pb.RegisterRequest) (*pb.RegisterReply, error) {

	newStudent := student{
		id:         getUUID(),
		name:       info.GetName(),
		age:        info.GetAge(),
		profession: info.GetProfession(),
	}
	allStudentInfo.mux.Lock()
	allStudentInfo.studentInfo[newStudent.id] = newStudent
	defer allStudentInfo.mux.Unlock()
	log.Printf("register %v success", newStudent.id)
	return &pb.RegisterReply{Id: newStudent.id}, nil
}

func (s *Server) Query(_ context.Context, studentId *pb.StudentInfo) (*pb.StudentInfo, error) {
	allStudentInfo.mux.Lock()
	studentInfo, ok := allStudentInfo.studentInfo[studentId.Id]
	defer allStudentInfo.mux.Unlock()
	if !ok {
		log.Print("student is not exist")
		return &pb.StudentInfo{}, errors.New("student is nor exist")
	}
	log.Printf("find student %v", studentId.Id)
	return &pb.StudentInfo{
		Id:         studentInfo.id,
		Name:       studentInfo.name,
		Age:        studentInfo.age,
		Profession: studentInfo.profession,
	}, nil
}

func (s *Server) AlterProfession(_ context.Context, alterInfo *pb.StudentInfo) (*pb.Result, error) {
	allStudentInfo.mux.Lock()
	studentInfo, ok := allStudentInfo.studentInfo[alterInfo.Id]
	defer allStudentInfo.mux.Unlock()
	if !ok {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	studentInfo.profession = alterInfo.Profession
	log.Printf("Alter student %v profession success", alterInfo.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) Delete(_ context.Context, studentId *pb.StudentInfo) (*pb.Result, error) {
	allStudentInfo.mux.Lock()
	_, ok := allStudentInfo.studentInfo[studentId.Id]
	defer allStudentInfo.mux.Unlock()
	if !ok {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	delete(allStudentInfo.studentInfo, studentId.Id)
	log.Printf("delete student %v success", studentId.Id)
	return &pb.Result{Res: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServiceServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

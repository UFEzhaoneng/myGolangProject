package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sort"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	pb "mygolangproject/proto"
)

const port = ":50052"

type Server struct {
	pb.UnimplementedServiceServer
}

type student struct {
	id           string //唯一
	name         string //仅支持英文，非空
	age          int32  //非空，范围【10，100】
	profession   string //枚举：计算机科学与技术/软件工程
	createTime   int64  //创建时间
	modifiedTime int64  //修改时间
}
type safeStudentInfo struct {
	studentInfo map[string]student
	mux         sync.RWMutex
}

var allStudentInfo = safeStudentInfo{studentInfo: make(map[string]student)}

func getUUID() string {
	return uuid.NewV4().String()
}

type studentList []student

func (stu studentList) Swap(i, j int)      { stu[i], stu[j] = stu[j], stu[i] }
func (stu studentList) Len() int           { return len(stu) }
func (stu studentList) Less(i, j int) bool { return stu[i].createTime > stu[j].createTime }

// A function to turn a map into a PairList, then sort and return it.
func sortByCreateTime(m map[string]student) studentList {
	p := make(studentList, len(m))
	i := 0
	for _, v := range m {
		p[i] = v
		i++
	}
	sort.Sort(p)
	return p
}

// SayHello implements helloworld.GreeterServer
func (s *Server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

//Register implements helloworld.GreeterServer
func (s *Server) Register(_ context.Context, info *pb.RegisterRequest) (*pb.RegisterReply, error) {

	newStudent := student{
		id:           getUUID(),
		name:         info.GetName(),
		age:          info.GetAge(),
		profession:   info.GetProfession(),
		createTime:   time.Now().Unix(),
		modifiedTime: time.Now().Unix(),
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
		return &pb.StudentInfo{}, errors.New("student is not exist")
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
	studentInfo.modifiedTime = time.Now().Unix()
	allStudentInfo.studentInfo[alterInfo.Id] = studentInfo
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

func (s *Server) QueryList(_ context.Context, _ *pb.QueryRequest) (*pb.StudentList, error) {
	allStudentInfo.mux.Lock()
	defer allStudentInfo.mux.Unlock()
	list := sortByCreateTime(allStudentInfo.studentInfo)
	studentList := &pb.StudentList{}
	for _, studentInfo := range list {
		studentInfo := &pb.StudentInfo{
			Id:           studentInfo.id,
			Name:         studentInfo.name,
			Age:          studentInfo.age,
			Profession:   studentInfo.profession,
			CreateTime:   studentInfo.createTime,
			ModifiedTime: studentInfo.modifiedTime,
		}
		studentList.StudentInfo = append(studentList.StudentInfo, studentInfo)
	}
	log.Print("query list success")
	return studentList, nil
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

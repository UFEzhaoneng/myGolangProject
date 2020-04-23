package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sort"
	"time"

	"google.golang.org/grpc"
	pb "mygolangproject/proto"
)

const port = ":50051"

type Server struct {
	pb.UnimplementedServiceServer
}

const (
	computerScienceAndTechnology = "计算机科学与技术"
	softwareEngineering          = "软件工程"
)

type student struct {
	id           int32  //唯一
	name         string //仅支持英文，非空
	age          int32  //非空，范围【10，100】
	profession   string //枚举：计算机科学与技术/软件工程
	createTime   int64  //创建时间
	modifiedTime int64  //修改时间
}

var allStudentInfo = make(map[int32]student)
var studentId = 0

type studentList []student

func (stu studentList) Swap(i, j int)      { stu[i], stu[j] = stu[j], stu[i] }
func (stu studentList) Len() int           { return len(stu) }
func (stu studentList) Less(i, j int) bool { return stu[i].createTime > stu[j].createTime }

// A function to turn a map into a PairList, then sort and return it.
func sortByCreateTime(m map[int32]student) studentList {
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
	if info.Profession != computerScienceAndTechnology && info.Profession != softwareEngineering {
		log.Printf("register error")
		return &pb.RegisterReply{}, errors.New("register error")
	}
	studentId++
	newStudent := student{
		id:           int32(studentId),
		name:         info.GetName(),
		age:          info.GetAge(),
		profession:   info.GetProfession(),
		createTime:   time.Now().Unix(),
		modifiedTime: time.Now().Unix(),
	}
	allStudentInfo[newStudent.id] = newStudent
	log.Printf("register %v success", newStudent.id)
	return &pb.RegisterReply{Id: newStudent.id}, nil
}

//查询学生信息
func (s *Server) Query(_ context.Context, studentId *pb.StudentId) (*pb.StudentInfo, error) {
	studentInfo, ok := allStudentInfo[studentId.Id]
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

//更改学生专业
func (s *Server) AlterProfession(_ context.Context, studentId *pb.StudentId) (*pb.Result, error) {
	studentInfo, ok := allStudentInfo[studentId.Id]
	if !ok {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	if studentInfo.profession == computerScienceAndTechnology {
		studentInfo.profession = softwareEngineering
	} else {
		studentInfo.profession = computerScienceAndTechnology
	}
	studentInfo.modifiedTime = time.Now().Unix()
	allStudentInfo[studentId.Id] = studentInfo
	log.Printf("Alter student %v profession success", studentId.Id)
	return &pb.Result{Res: true}, nil
}

//删除学生信息
func (s *Server) Delete(_ context.Context, studentId *pb.StudentId) (*pb.Result, error) {
	_, ok := allStudentInfo[studentId.Id]
	if !ok {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	delete(allStudentInfo, studentId.Id)
	log.Printf("delete student %v success", studentId.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) QueryList(_ context.Context, _ *pb.QueryRequest) (*pb.StudentList, error) {
	list := sortByCreateTime(allStudentInfo)
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

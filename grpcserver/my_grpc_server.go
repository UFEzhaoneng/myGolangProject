package main

import (
	"context"
	"errors"
	"hash/crc32"
	"log"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	pb "mygolangproject/proto"
)

const port = ":50052"

type Server struct {
	pb.UnimplementedServiceServer
}

func connectMysql() (*gorm.DB, error) {
	db, err := gorm.Open("mysql", "root:123456@/my_golang_project?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Print("connect mysql error!")
		return nil, err
	}
	if !db.HasTable("students") {
		if err := db.CreateTable(&student{}).Error; err != nil {
			log.Print(err)
			return nil, errors.New("create table error")
		}
	}
	return db, err
}

type student struct {
	ID         uint
	Name       string //仅支持英文，非空
	Age        int32  //非空，范围【10，100】
	Profession string //枚举：计算机科学与技术/软件工程
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func getUUID() uint32 {
	v4 := uuid.NewV4()
	return crc32.ChecksumIEEE([]byte(v4.String()))
}

type studentList []student

func (stu studentList) Swap(i, j int)      { stu[i], stu[j] = stu[j], stu[i] }
func (stu studentList) Len() int           { return len(stu) }
func (stu studentList) Less(i, j int) bool { return stu[i].CreatedAt.Unix() > stu[j].CreatedAt.Unix() }

// A function to turn a map into a PairList, then sort and return it.
func sortByCreateTime(m map[int]student) studentList {
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

	db, err := connectMysql()
	defer db.Close()
	if err != nil {
		return nil, err
	}
	newStudent := student{
		ID:         uint(getUUID()),
		Name:       info.GetName(),
		Age:        info.GetAge(),
		Profession: info.GetProfession(),
	}
	db.Create(&newStudent)
	log.Printf("register %v success", newStudent.ID)
	return &pb.RegisterReply{Id: strconv.Itoa(int(newStudent.ID))}, nil
}

func studentQuery(id string) (student, *gorm.DB) {
	db, err := connectMysql()
	newStudent := student{}
	if err != nil {
		return newStudent, db
	}
	db.Where("id = ?", id).First(&newStudent)
	return newStudent, db
}

func (s *Server) Query(_ context.Context, studentId *pb.StudentInfo) (*pb.StudentInfo, error) {
	newStudent, db := studentQuery(studentId.Id)
	defer db.Close()
	if newStudent.ID == 0 {
		return &pb.StudentInfo{}, errors.New("student is not exist")
	}
	log.Printf("find student %v", studentId.Id)
	return &pb.StudentInfo{
		Id:         strconv.Itoa(int(newStudent.ID)),
		Name:       newStudent.Name,
		Age:        newStudent.Age,
		Profession: newStudent.Profession,
	}, nil
}

func (s *Server) AlterProfession(_ context.Context, alterInfo *pb.StudentInfo) (*pb.Result, error) {
	newStudent, db := studentQuery(alterInfo.Id)
	defer db.Close()
	if newStudent.ID == 0 {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}

	db.Model(&newStudent).Update("profession", alterInfo.Profession)
	log.Printf("Alter student %v Profession success", alterInfo.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) Delete(_ context.Context, studentId *pb.StudentInfo) (*pb.Result, error) {
	newStudent, db := studentQuery(studentId.Id)
	defer db.Close()
	if newStudent.ID == 0 {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	db.Delete(&newStudent)
	log.Printf("delete student %v success", studentId.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) QueryList(_ context.Context, _ *pb.QueryRequest) (*pb.StudentList, error) {
	studentList := &pb.StudentList{}
	db, err := connectMysql()
	if err != nil {
		return nil, err
	}
	var students []student
	defer db.Close()
	db.Find(&students)
	for _, studentInfo := range students {
		studentInfo := &pb.StudentInfo{
			Id:           strconv.Itoa(int(studentInfo.ID)),
			Name:         studentInfo.Name,
			Age:          studentInfo.Age,
			Profession:   studentInfo.Profession,
			CreateTime:   studentInfo.CreatedAt.Unix(),
			ModifiedTime: studentInfo.UpdatedAt.Unix(),
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

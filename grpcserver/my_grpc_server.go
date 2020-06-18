package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	uuid "github.com/satori/go.uuid"
	"mygolangproject/consul"
	pb "mygolangproject/proto"
)

const port = ":50052"

var count int64

var db = connectMysql()

type Server struct {
	pb.UnimplementedServiceServer
}

func register() {

	defer func() {
		fmt.Println("启动错误")
	}()
	//使用consul注册服务
	register := consul.NewConsulRegister()
	register.Port = 50052
	register.Name = "grpcServer"
	register.Tag = []string{"grpc"}
	if err := register.GRPCRegister(); err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(s, &consul.HealthImpl{Status: grpc_health_v1.HealthCheckResponse_SERVING})
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("grpc server ready listen")
	pb.RegisterServiceServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func connectMysql() *gorm.DB {
	db, err := gorm.Open("mysql", "root:123456@/my_golang_project?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("connect mysql error!")
		return nil
	}
	if !db.HasTable("students") {
		log.Fatal("table not exist!")
		return nil
	}
	return db
}

type student struct {
	ID         string //唯一
	Name       string //仅支持英文，非空
	Age        int32  //非空，范围【10，100】
	Profession string //枚举：计算机科学与技术/软件工程
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func getUUID() string {
	return uuid.NewV4().String()
}

// SayHello implements helloworld.GreeterServer
func (s *Server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

//GRPCRegister implements helloworld.GreeterServer
func (s *Server) Register(_ context.Context, info *pb.RegisterRequest) (*pb.RegisterReply, error) {
	newStudent := student{
		ID:         getUUID(),
		Name:       info.GetName(),
		Age:        info.GetAge(),
		Profession: info.GetProfession(),
	}
	tx := db.Begin()
	if err := tx.Create(&newStudent).Error; err != nil {
		tx.Rollback()
		log.Print(err)
		return nil, err
	} else {
		tx.Commit()
	}
	log.Printf("register %v success", newStudent.ID)
	return &pb.RegisterReply{Id: newStudent.ID}, nil
}

func studentQuery(id string) student {
	newStudent := student{}
	tx := db.Begin()
	if err := tx.Where("id = ?", id).First(&newStudent).Error; err != nil {
		tx.Rollback()
		log.Print(err)
	} else {
		tx.Commit()
	}
	return newStudent
}

func (s *Server) Query(_ context.Context, studentId *pb.StudentInfo) (*pb.StudentInfo, error) {
	newStudent := studentQuery(studentId.Id)
	if newStudent.ID == "" {
		return &pb.StudentInfo{}, errors.New("student is not exist")
	}
	log.Printf("find student %v", studentId.Id)
	return &pb.StudentInfo{
		Id:         newStudent.ID,
		Name:       newStudent.Name,
		Age:        newStudent.Age,
		Profession: newStudent.Profession,
	}, nil
}

func (s *Server) AlterProfession(_ context.Context, alterInfo *pb.StudentInfo) (*pb.Result, error) {
	newStudent := studentQuery(alterInfo.Id)
	if newStudent.ID == "" {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is not exist")
	}
	tx := db.Begin()
	if err := tx.Model(&newStudent).Update("profession", alterInfo.Profession).Error; err != nil {
		db.Rollback()
		log.Print(err)
		return &pb.Result{Res: false}, err
	} else {
		tx.Commit()
	}
	log.Printf("Alter student %v Profession success", alterInfo.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) Delete(_ context.Context, studentId *pb.StudentInfo) (*pb.Result, error) {
	newStudent := studentQuery(studentId.Id)
	if newStudent.ID == "" {
		log.Print("student is not exist")
		return &pb.Result{Res: false}, errors.New("student is nor exist")
	}
	tx := db.Begin()
	if err := tx.Delete(&newStudent).Error; err != nil {
		tx.Rollback()
		log.Print(err)
		return &pb.Result{Res: false}, err
	} else {
		tx.Commit()
	}
	log.Printf("delete student %v success", studentId.Id)
	return &pb.Result{Res: true}, nil
}

func (s *Server) QueryList(_ context.Context, _ *pb.QueryRequest) (*pb.StudentList, error) {
	studentList := &pb.StudentList{}
	var students []student
	tx := db.Begin()
	if err := tx.Find(&students).Error; err != nil {
		tx.Rollback()
		log.Print(err)
		return studentList, err
	} else {
		tx.Commit()
	}
	for _, studentInfo := range students {
		studentInfo := &pb.StudentInfo{
			Id:           studentInfo.ID,
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
	register()
}

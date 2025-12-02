package main

import (
	"context"
	"fmt"

	"github.com/khunfloat/irpc"
	"github.com/khunfloat/irpc/example/client"
	"github.com/khunfloat/irpc/example/contract"
)

// --------------------------
// Implementation (Service)
// --------------------------

type Exam struct {
	Id   string
	Name string
}

type ExamService interface {
	FindExamById(ctx context.Context, id string) (*Exam, error)
	FindAllExams(ctx context.Context) ([]*Exam, error)
}

type examService struct{}

func NewExamService() ExamService {
	return &examService{}
}

func (s *examService) FindExamById(ctx context.Context, id string) (*Exam, error) {
	// mock data
	return &Exam{
		Id:   id,
		Name: "Mock Exam",
	}, nil
}

func (s *examService) FindAllExams(ctx context.Context) ([]*Exam, error) {
	var exams []*Exam
	// mock data
	exams = append(exams, &Exam{Id: "EX-001", Name: "Exam 1"})
	exams = append(exams, &Exam{Id: "EX-002", Name: "Exam 2"})
	exams = append(exams, &Exam{Id: "EX-003", Name: "Exam 3"})

	return exams, nil
}

// --------------------------
// Implement Contract Interface
// --------------------------

type ExamInterface interface {
	FindExamById(ctx context.Context, req contract.ExamContractReq) (*contract.ExamContractRes, error)
	FindAllExams(ctx context.Context) ([]*contract.ExamContractRes, error)
}

type examInterface struct {
	service ExamService
}

func (i *examInterface) FindExamById(ctx context.Context, req contract.ExamContractReq) (*contract.ExamContractRes, error) {
	exam, err := i.service.FindExamById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &contract.ExamContractRes{
		Id:   exam.Id,
		Name: exam.Name,
	}, nil
}

func (i *examInterface) FindAllExams(ctx context.Context) ([]*contract.ExamContractRes, error) {
	exams, err := i.service.FindAllExams(ctx)
	if err != nil {
		return nil, err
	}

	var res []*contract.ExamContractRes
	for _, e := range exams {
		res = append(res, &contract.ExamContractRes{
			Id:   e.Id,
			Name: e.Name,
		})
	}

	return res, nil
}

// --------------------------
// Main: test IRPC
// --------------------------

func main() {
	ctx := context.Background()

	// create registry
	// with default config: AllowOverride = false
	// AllowOverride can be set to true if you want to allow method override during registration
	registry := irpc.NewRegistry(irpc.DEFAULT_CONFIG)

	// create service
	examService := NewExamService()

	// create contract implementation
	examInterface := &examInterface{
		service: examService,
	}

	// register
	registry.RegisterContract("Exam", (*contract.ExamContract)(nil), examInterface)

	// irpc client
	examClient := client.NewExamClient(registry)

	// -------- test call: FindExamById --------
	res1, err := examClient.FindExamById(ctx, contract.ExamContractReq{Id: "EX-123"})
	if err != nil {
		panic(err)
	}

	fmt.Println("FindExamById:", res1.Id, res1.Name)

	// -------- test call: FindAllExams --------
	res2, err := examClient.FindAllExams(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("FindAllExams:")
	for _, e := range res2 {
		fmt.Println("-", e.Id, e.Name)
	}
}

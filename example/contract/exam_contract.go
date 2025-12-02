package contract

import "context"

type ExamContractReq struct {
	Id string
}

type ExamContractRes struct {
	Id   string
	Name string
}

type ExamContract interface {
	FindExamById(ctx context.Context, req ExamContractReq) (*ExamContractRes, error)
	FindAllExams(ctx context.Context) ([]ExamContractRes, error)
}

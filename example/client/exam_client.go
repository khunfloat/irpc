package client

import (
	"context"

	"github.com/khunfloat/irpc"
	"github.com/khunfloat/irpc/example/contract"
)

type ExamClient interface {
	FindExamById(ctx context.Context, req contract.ExamContractReq) (*contract.ExamContractRes, error)
	FindAllExams(ctx context.Context) ([]*contract.ExamContractRes, error)
}

type examClient struct {
	registry *irpc.Registry
}

func NewExamClient(registry *irpc.Registry) ExamClient {
	return &examClient{
		registry: registry,
	}
}

func (c *examClient) FindExamById(ctx context.Context, req contract.ExamContractReq) (*contract.ExamContractRes, error) {
	res, err := c.registry.Call(ctx, "Exam.FindExamById", req)
	if err != nil {
		return nil, err
	}
	return res.(*contract.ExamContractRes), nil
}

func (c *examClient) FindAllExams(ctx context.Context) ([]*contract.ExamContractRes, error) {
	res, err := c.registry.Call(ctx, "Exam.FindAllExams", nil)
	if err != nil {
		return nil, err
	}
	return res.([]*contract.ExamContractRes), nil
}

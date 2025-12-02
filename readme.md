# **IRPC — Lightweight In-Process RPC for Go**

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go">
  <img src="https://img.shields.io/badge/License-MIT-green.svg">
  <img src="https://img.shields.io/github/v/release/khunfloat/irpc">
</p>

IRPC is a minimal and fast in-process RPC system for Go applications.
It lets modules call each other using RPC-style method names without using gRPC, HTTP, or any network transport.
Everything runs in memory with near-zero overhead.

- No network cost
- No reflection at call time
- Automatic contract (interface) registration
- Clean Architecture friendly
- Extremely small API surface
- Simple, predictable behavior

## **Features**

- Register all methods from an interface as RPC endpoints
- Zero-reflection on method invocation (high performance)
- Optional override mode when re-registering keys
- Uses interfaces as contracts (similar to gRPC service definitions)
- Easy to build client stubs on top of the Registry

## **Installation**

```sh
go get github.com/khunfloat/irpc
```

## **Example Overview**

This example demonstrates a full flow:

1. Define a **contract interface**
2. Implement the interface using your domain service
3. Register the contract with IRPC
4. Create a simple client
5. Call methods via IRPC

Directory structure:

```
example/
 ├── contract/
 ├── client/
 └── main.go
```

## **1. Define a Contract (Interface)**

```go
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
```

This interface serves as the RPC definition for the module.

## **2. Implement the Contract**

```go
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
```

### Contract implementation wrapper:

```go
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
```

## **3. Register the Contract with IRPC**

```go
registry := irpc.NewRegistry(irpc.DEFAULT_CONFIG)

examService := NewExamService()
examInterface := &examInterface{service: examService}

registry.RegisterContract("Exam", (*contract.ExamContract)(nil), examInterface)
```

This registers:

```
Exam.FindExamById
Exam.FindAllExams
```

## **4. Create a Client Wrapper**

```go
package client

type examClient struct {
	registry *irpc.Registry
}

func (c *examClient) FindExamById(ctx context.Context, req contract.ExamContractReq) (*contract.ExamContractRes, error) {
	res, err := c.registry.Call(ctx, "Exam.FindExamById", req)
	if err != nil {
		return nil, err
	}
	return res.(*contract.ExamContractRes), nil
}
```

## **5. Call Methods via IRPC**

```go
res1, err := examClient.FindExamById(ctx, contract.ExamContractReq{Id: "EX-123"})
fmt.Println("FindExamById:", res1.Id, res1.Name)
```

Output:

```
FindExamById: EX-123 Mock Exam
```

# **IRPC Registry Logic**

Below is the core implementation included in this project:

```go
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
	config   Config
}
```

### Register Contract

```go
func (r *Registry) RegisterContract(serviceName string, iface any, impl any)
```

This:

- Reads all methods from the interface
- Ensures the implementation implements them
- Creates a fast invocation wrapper
- Registers keys such as `Exam.FindExamById`

### Call Method

```go
func (r *Registry) Call(ctx context.Context, key string, req any) (any, error)
```

This performs a constant-time lookup and calls the handler without reflection.

## **Configuration**

```go
type Config struct {
	AllowOverride bool
}

var DEFAULT_CONFIG = Config{
	AllowOverride: false,
}
```

### Examples

**Safe mode (default):**

```go
registry := irpc.NewRegistry(irpc.DEFAULT_CONFIG)
```

Duplicate keys → panic.

**Allow override:**

```go
registry := irpc.NewRegistry(irpc.Config{AllowOverride: true})
```

Duplicate keys → overwritten silently.

## **Why IRPC?**

- Perfect for monoliths with modular architecture
- Eliminates boilerplate between layers
- Safer than using `interface{}`-driven
- Encourages clean & explicit service contracts

If you are using Clean Architecture or Hexagonal Architecture, IRPC fits naturally as the "in-process transport layer."

## **License**

- MIT

## **Contributors**

- ChatGPT (Most Everything)
- Kositwanich C. (Just Prompting)

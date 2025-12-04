/*
Package irpc provides a lightweight, in-process RPC mechanism for Go applications.

It allows modules to invoke each other using RPC-style keys without using
network transports such as HTTP or gRPC. IRPC is designed to be minimal,
fast, and suitable for Clean Architecture and Hexagonal Architecture.

Key Concepts

  - A "contract" is an interface that defines the available RPC methods.
  - An "implementation" is a struct that implements the contract.
  - IRPC automatically registers all methods from the interface with a prefix.
  - Methods are invoked through a fast, reflection-free handler.

Example Usage

    registry := irpc.NewRegistry(irpc.DEFAULT_CONFIG)

    // Register a contract
    registry.RegisterContract("Exam", (*ExamContract)(nil), examService)

    // Call a method
    res, err := registry.Call(ctx, "Exam.FindExamById", ExamRequest{Id: "EX-1"})
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(res.(*ExamResponse).Name)

# Registry

NewRegistry(config Config) *Registry

    Creates a new registry with the provided configuration. If AllowOverride
    is false, registering the same key twice will produce a panic.

RegisterContract(serviceName string, iface any, impl any)

    Registers all methods declared in the given interface (iface) and binds
    them to the implementation (impl). Each method is registered under the key:
        serviceName + "." + MethodName

Register(key string, h HandlerFunc)

    Registers a handler function for a specific RPC key.

Call(ctx context.Context, key string, req any)

    Invokes a registered handler. Panics or returns an error if the key does
    not exist.

# Configuration

    type Config struct {
        AllowOverride bool
        AllowPartial  bool
    }

    var DEFAULT_CONFIG = Config{
        AllowOverride: false,
        AllowPartial:  false,
    }

If AllowPartial is true, RegisterContract will silently skip missing methods
instead of panicking.

HandlerFunc

    type HandlerFunc func(ctx context.Context, req any) (any, error)

Performance Characteristics remain unchanged.

*/

package irpc

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Config struct {
	AllowOverride bool
	AllowPartial  bool
}

var DEFAULT_CONFIG = Config{
	AllowOverride: false,
	AllowPartial:  false,
}

type HandlerFunc func(context.Context, any) (any, error)

type Registry struct {
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
	config   Config
}

func NewRegistry(config Config) *Registry {
	return &Registry{
		handlers: make(map[string]HandlerFunc),
		config:   config,
	}
}

func (r *Registry) RegisterContract(serviceName string, iface any, impl any) {
	ifaceType := reflect.TypeOf(iface).Elem()
	implVal := reflect.ValueOf(impl)
	implType := implVal.Type()

	if implType.Kind() != reflect.Pointer {
		panic("irpc: impl must be a pointer to struct")
	}

	for i := 0; i < ifaceType.NumMethod(); i++ {
		ifaceMethod := ifaceType.Method(i)
		mName := ifaceMethod.Name

		implMethod := implVal.MethodByName(mName)
		if !implMethod.IsValid() {
			if r.config.AllowPartial {
				continue
			}
			panic(fmt.Sprintf("irpc: missing method: %s.%s", serviceName, mName))
		}

		key := serviceName + "." + mName
		if _, exists := r.handlers[key]; exists && !r.config.AllowOverride {
			panic(fmt.Sprintf("irpc: duplicate method key '%s' in RegisterContract", key))
		}

		h := makeHandler(implMethod)

		r.Register(key, h)
	}
}

func makeHandler(method reflect.Value) HandlerFunc {
	return func(ctx context.Context, req any) (any, error) {
		in := []reflect.Value{reflect.ValueOf(ctx)}

		if method.Type().NumIn() == 2 {
			in = append(in, reflect.ValueOf(req))
		}

		out := method.Call(in)

		var err error
		if len(out) == 2 && !out[1].IsNil() {
			err = out[1].Interface().(error)
		}

		if len(out) >= 1 {
			return out[0].Interface(), err
		}
		return nil, err
	}
}

func (r *Registry) Register(key string, h HandlerFunc) {
	r.mu.Lock()
	r.handlers[key] = h
	r.mu.Unlock()
}

func (r *Registry) Call(ctx context.Context, key string, req any) (any, error) {
	r.mu.RLock()
	h := r.handlers[key]
	r.mu.RUnlock()

	if h == nil {
		return nil, fmt.Errorf("irpc: handler not found: %s", key)
	}

	return h(ctx, req)
}

func (r *Registry) ValidateImpl(serviceName string, iface any) {
	ifaceType := reflect.TypeOf(iface).Elem()

	for i := 0; i < ifaceType.NumMethod(); i++ {
		mName := ifaceType.Method(i).Name
		key := serviceName + "." + mName

		r.mu.RLock()
		_, exists := r.handlers[key]
		r.mu.RUnlock()

		if !exists {
			panic(fmt.Sprintf("irpc: missing registered handler for %s", key))
		}
	}
}

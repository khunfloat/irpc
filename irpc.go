package irpc

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Config struct {
	AllowOverride bool
}

var DEFAULT_CONFIG = Config{
	AllowOverride: false,
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

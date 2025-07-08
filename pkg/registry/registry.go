package Registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

type Middleware_Function func(Entry_Data) error

// function_registry stores dynamically registered functions.
type Function_Registry struct {
	func_map               map[string]reflect.Value
	middleware_list        []Middleware_Function
	i_username, i_password string
	mu                     sync.RWMutex
}

type Entry_Data struct {
	Bulk_Data    interface{}
	Http_Request *http.Request
}

// new_registry creates a new registry.
func New_Registry() *Function_Registry {
	return &Function_Registry{
		func_map: make(map[string]reflect.Value),
	}
}

// registers a function with a specific name.
func (r *Function_Registry) Add(name string, function interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	fn_value := reflect.ValueOf(function)
	if fn_value.Kind() != reflect.Func {
		return errors.New("provided value is not a function")
	}
	r.func_map[name] = fn_value
	return nil
}

// convert_arg tries to convert the received value (from JSON) to the target type.
func convert_arg(arg interface{}, target_type reflect.Type) (reflect.Value, error) {
	if arg == nil {
		return reflect.Zero(target_type), nil
	}

	arg_value := reflect.ValueOf(arg)
	if arg_value.Type().AssignableTo(target_type) {
		return arg_value, nil
	}

	// In JSON, numbers are represented as float64.
	switch target_type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := arg.(float64); ok {
			return reflect.ValueOf(int(num)).Convert(target_type), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := arg.(float64); ok {
			return reflect.ValueOf(uint(num)).Convert(target_type), nil
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := arg.(float64); ok {
			return reflect.ValueOf(num).Convert(target_type), nil
		}
	case reflect.String:
		if s, ok := arg.(string); ok {
			return reflect.ValueOf(s), nil
		}
	case reflect.Bool:
		if b, ok := arg.(bool); ok {
			return reflect.ValueOf(b), nil
		}
	}

	// Attempt to marshal/unmarshal for complex structures.
	new_val := reflect.New(target_type)
	bytes, err := json.Marshal(arg)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to marshal argument: %v", err)
	}
	err = json.Unmarshal(bytes, new_val.Interface())
	if err != nil {
		return reflect.Value{}, fmt.Errorf("failed to unmarshal argument to %s: %v", target_type, err)
	}

	if target_type.Kind() != reflect.Ptr {
		return new_val.Elem(), nil
	}
	return new_val, nil
}

// call executes a registered function.
func (r *Function_Registry) Call(ctx context.Context, name string, args ...interface{}) ([]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, exists := r.func_map[name]
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	fn_type := fn.Type()
	num_in := fn_type.NumIn()
	arg_offset := 0
	if num_in > 0 && fn_type.In(0) == reflect.TypeOf((*context.Context)(nil)).Elem() {
		arg_offset = 1
	}

	if len(args) != num_in-arg_offset {
		return nil, fmt.Errorf("expected %d arguments, got %d", num_in-arg_offset, len(args))
	}

	in := make([]reflect.Value, num_in)
	if arg_offset == 1 {
		in[0] = reflect.ValueOf(ctx)
	}
	for i := 0; i < len(args); i++ {
		expected_type := fn_type.In(i + arg_offset)
		converted_arg, err := convert_arg(args[i], expected_type)
		if err != nil {
			return nil, fmt.Errorf("argument %d conversion error: %v", i+1, err)
		}
		if !converted_arg.Type().AssignableTo(expected_type) {
			return nil, fmt.Errorf("argument %d must be %s", i+1, expected_type)
		}
		in[i+arg_offset] = converted_arg
	}

	results := fn.Call(in)
	output := make([]interface{}, len(results))
	for i, result := range results {
		output[i] = result.Interface()
	}
	return output, nil
}

// Register a middleware function.
func (r *Function_Registry) Use(handler Middleware_Function) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware_list = append(r.middleware_list, handler)
}

func (r *Function_Registry) Invoke_Middlewares(data []struct {
	ID        interface{}   `json:"id"`
	Func_Name string        `json:"func"`
	Args      []interface{} `json:"args"`
}, req *http.Request) error {
	entry_data := Entry_Data{Bulk_Data: data, Http_Request: req}
	for _, function := range r.middleware_list {
		err := function(entry_data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Function_Registry) Set_Basic_Auth(username string, password string) {
	r.i_username = username
	r.i_password = password
}

func (r *Function_Registry) Check_Authentication(req *http.Request) (bool, error) {
	if r.i_username == "" || r.i_password == "" {
		return true, nil
	}
	if username, password, ok := req.BasicAuth(); ok == true {
		if username == r.i_username && password == r.i_password {
			return true, nil
		} else {
			return false, errors.New("Invalid username or password")
		}

	}
	return false, errors.New("Invalid authorization format")
}

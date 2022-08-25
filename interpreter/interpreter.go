package interpreter

import (
	"context"
	"encoding/json"
	"github.com/fatih/color"
	"github.com/lusory/libkitsune"
	"reflect"
	"strings"
)

// Interpreter is an implementation of InterpreterInterface responsible for interpreting
// interactive console commands.
type Interpreter struct {
	Client    libkitsune.KitsuneClient
	logErrors bool
}

// InterpreterInterface is the abstract version of Interpreter.
type InterpreterInterface interface {
	Interpret(cmd string) string
}

// NewInterpreter connects to the supplied gRPC target and creates an Interpreter instance with it.
func NewInterpreter(target string, ssl bool, logErrors bool) (*Interpreter, error) {
	client, err := libkitsune.NewKitsuneClient(target, ssl)
	if err != nil {
		return &Interpreter{}, err
	}
	return &Interpreter{client, logErrors}, nil
}

// Interpret interprets an interactive console command and returns the output.
func (i *Interpreter) Interpret(cmd string) string {
	params := strings.SplitN(cmd, " ", 3)
	if len(params) < 1 || len(params) > 2 {
		if i.logErrors {
			color.Red("Invalid syntax: <registry>.<method> [data]")
		}
		return ""
	}

	methodSplit := strings.SplitN(params[0], ".", 2)
	if len(methodSplit) != 2 {
		if i.logErrors {
			color.Red("Invalid syntax for first parameter: <registry>.<method>")
		}
		return ""
	}

	var registry reflect.Value
	switch methodSplit[0] {
	case "img":
		registry = reflect.ValueOf(i.Client.ImageRegistry)
	case "vm":
		registry = reflect.ValueOf(i.Client.VmRegistry)
	default:
		if i.logErrors {
			color.Red("Invalid registry, must be image or vm")
		}
		return ""
	}

	method := registry.MethodByName(methodSplit[1])
	if !method.IsValid() {
		if i.logErrors {
			color.Red("Invalid method for %s registry", methodSplit[0])
		}
		return ""
	}

	// get second method parameter, depointerize type and instantiate struct
	data := reflect.New(method.Type().In(1).Elem()).Interface()

	// default data is {}
	jsonData := "{}"
	if len(params) == 2 {
		jsonData = params[1]
	}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		if i.logErrors {
			color.Red("Invalid data, JSON deserialization error: %s", err.Error())
		}
		return ""
	}

	returnParams := method.Call([]reflect.Value{
		reflect.ValueOf(context.Background()),
		reflect.ValueOf(data),
	})

	// check if gRPC returned error
	if err := returnParams[1].Interface(); err != nil {
		if i.logErrors {
			color.Red("gRPC error: %s", err.(error).Error())
		}
		return ""
	}

	// marshal depointerized gRPC result to JSON
	out, err := json.Marshal(returnParams[0].Elem().Interface())
	if err != nil {
		if i.logErrors {
			color.Red("JSON marshalling error: %s", err.Error())
		}
		return ""
	}
	return string(out)
}

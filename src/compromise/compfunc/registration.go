package compfunc

import (
	"fmt"
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/go-common/src/common"
	"reflect"
	"strings"
)

// These are the function types that are allowed to be registered.
type (
	simpleFunc                          = func()
	oneArgFunc                          = func(arg string)
	varArgFunc                          = func(args []string)
	simpleFuncWithContext               = func(context compromise.CompleteContext)
	oneArgFuncWithContext               = func(context compromise.CompleteContext, arg string)
	varArgFuncWithContext               = func(context compromise.CompleteContext, args []string)
	simpleCandidateGenerator            = func() compromise.CandidateList
	oneArgCandidateGenerator            = func(arg string) compromise.CandidateList
	varArgCandidateGenerator            = func(args []string) compromise.CandidateList
	simpleCandidateGeneratorWithContext = func(context compromise.CompleteContext) compromise.CandidateList
	oneArgCandidateGeneratorWithContext = func(context compromise.CompleteContext, arg string) compromise.CandidateList
	varArgCandidateGeneratorWithContext = func(context compromise.CompleteContext, args []string) compromise.CandidateList
)

func getFunctionAdapter(function interface{}, name string) (varArgCandidateGeneratorWithContext, error) {
	firstArg := func(args []string) string {
		if len(args) > 0 {
			return args[0]
		}
		return ""
	}

	if f, ok := function.(simpleFunc); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f()
			return nil
		}, nil
	}
	if f, ok := function.(oneArgFunc); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f(firstArg(args))
			return nil
		}, nil
	}
	if f, ok := function.(varArgFunc); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f(args)
			return nil
		}, nil
	}
	if f, ok := function.(simpleFuncWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f(context)
			return nil
		}, nil
	}
	if f, ok := function.(oneArgFuncWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f(context, firstArg(args))
			return nil
		}, nil
	}
	if f, ok := function.(varArgFuncWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			f(context, args)
			return nil
		}, nil
	}
	if f, ok := function.(simpleCandidateGenerator); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f()
		}, nil
	}
	if f, ok := function.(oneArgCandidateGenerator); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f(firstArg(args))
		}, nil
	}
	if f, ok := function.(varArgCandidateGenerator); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f(args)
		}, nil
	}
	if f, ok := function.(simpleCandidateGeneratorWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f(context)
		}, nil
	}
	if f, ok := function.(oneArgCandidateGeneratorWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f(context, firstArg(args))
		}, nil
	}
	if f, ok := function.(varArgCandidateGeneratorWithContext); ok {
		return func(context compromise.CompleteContext, args []string) compromise.CandidateList {
			return f(context, args)
		}, nil
	}
	return nil, fmt.Errorf("invalid signature of function %s: %v", name, reflect.TypeOf(function))
}

var (
	// All registered functions.
	funcs = make(map[string]varArgCandidateGeneratorWithContext)
)

// Register registers a new callback function associated with a given name.
// Only function types defined in this file are allowed.
func Register(name string, function interface{}) {
	if common.DebugEnabled {
		common.Debugf("Registering function: name=%s value=%v type=%v", name, function, reflect.TypeOf(function))
	}
	if len(name) == 0 {
		panic("function name must not be empty")
	}
	lname := strings.ToLower(name)
	if _, ok := funcs[lname]; ok {
		panic(fmt.Sprintf("function \"%s\" already defined", name))
	}
	if function == nil {
		panic("function cannot be nil")
	}
	adapter, err := getFunctionAdapter(function, name)
	if err != nil {
		panic(err.Error())
	}
	funcs[lname] = adapter
}

// Defined returns whether a function with a given name is registered.
func Defined(name string) error {
	_, err := getFunction(name)
	return err
}

func getFunction(name string) (varArgCandidateGeneratorWithContext, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("function name must not be empty")
	}
	lname := strings.ToLower(name)
	if adapter, ok := funcs[lname]; ok {
		return adapter, nil
	}
	return nil, fmt.Errorf("function \"%s\" not defined", name)
}

// Invoke invokes a registered function.
func Invoke(name string, context compromise.CompleteContext, args []string) compromise.CandidateList {
	adapter, err := getFunction(name)
	common.CheckPanice(err) // The function name must have been verified already, so let's panic.

	return adapter(context, args)
}

package forceexport

import (
	go114 "github.com/AlaxLee/go-forceexport/go114"
	go116 "github.com/AlaxLee/go-forceexport/go116"
	"regexp"
	"runtime"
)

var GoVersion string

func init() {
	reg := regexp.MustCompile(`go\d+\.\d+`)
	GoVersion = reg.FindString(runtime.Version())
}

// GetFunc gets the function defined by the given fully-qualified name. The
// outFuncPtr parameter should be a pointer to a function with the appropriate
// type (e.g. the address of a local variable), and is set to a new function
// value that calls the specified function. If the specified function does not
// exist, outFuncPtr is not set and an error is returned.
func GetFunc(outFuncPtr interface{}, name string) (err error) {
	switch GoVersion {
	case "go1.14", "go1.15":
		err = go114.GetFunc(outFuncPtr, name)
	case "go1.16":
		err = go116.GetFunc(outFuncPtr, name)
	default:
		panic("Not suitable for " + GoVersion)
	}
	return
}

// CreateFuncForCodePtr is given a code pointer and creates a function value
// that uses that pointer. The outFun argument should be a pointer to a function
// of the proper type (e.g. the address of a local variable), and will be set to
// the result function value.
func CreateFuncForCodePtr(outFuncPtr interface{}, codePtr uintptr) {
	switch GoVersion {
	case "go1.14", "go1.15":
		go114.CreateFuncForCodePtr(outFuncPtr, codePtr)
	case "go1.16":
		go116.CreateFuncForCodePtr(outFuncPtr, codePtr)
	default:
		panic("Not suitable for " + GoVersion)
	}
}

// FindFuncWithName searches through the moduledata table created by the linker
// and returns the function's code pointer. If the function was not found, it
// returns an error. Since the data structures here are not exported, we copy
// them below (and they need to stay in sync or else things will fail
// catastrophically).
func FindFuncWithName(name string) (p uintptr, err error) {
	switch GoVersion {
	case "go1.14", "go1.15":
		p, err = go114.FindFuncWithName(name)
	case "go1.16":
		p, err = go116.FindFuncWithName(name)
	default:
		panic("Not suitable for " + GoVersion)
	}
	return
}

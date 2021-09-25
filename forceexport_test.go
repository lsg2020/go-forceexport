package forceexport

import (
	"reflect"
	"runtime"
	"testing"
)

func FuncName(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func aTestTimeNow(t *testing.T) {
	var timeNowFunc func() (int64, int32)
	_ = GetFunc(&timeNowFunc, "time.now")
	sec, nsec := timeNowFunc()
	if sec == 0 || nsec == 0 {
		t.Error("Expected nonzero result from time.now().")
	}
}

// Note that we need to disable inlining here, or else the function won't be
// compiled into the binary. We also need to call it from the test so that the
// compiler doesn't remove it because it's unused.
//go:noinline
func addOne(x int) int {
	return x + 1
}

func aTestAddOne(t *testing.T) {
	if addOne(3) != 4 {
		t.Error("addOne should work properly.")
	}

	var addOneFunc func(x int) int
	err := GetFunc(&addOneFunc, "github.com/lsg2020/go-forceexport.addOne")
	if err != nil {
		t.Error("Expected nil error.")
	}
	if addOneFunc(3) != 4 {
		t.Error("Expected addOneFunc to add one to 3.")
	}
}

func aTestGetSelf(t *testing.T) {
	var getFunc func(interface{}, string) error
	err := GetFunc(&getFunc, "github.com/lsg2020/go-forceexport.GetFunc")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	// The two functions should share the same code pointer, so they should
	// have the same string representation.
	if FuncName(getFunc) != FuncName(GetFunc) {
		t.Errorf("Expected ")
	}
	// Call it again on itself!
	err = getFunc(&getFunc, "github.com/lsg2020/go-forceexport.GetFunc")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if FuncName(getFunc) != FuncName(GetFunc) {
		t.Errorf("Expected ")
	}
}

func TestInvalidFunc(t *testing.T) {
	var invalidFunc func()
	err := GetFunc(&invalidFunc, "invalidpackage.invalidfunction")
	if err == nil {
		t.Error("Expected an error.")
	}
	if invalidFunc != nil {
		t.Error("Expected a nil function.")
	}
}

// BenchmarkGetMain check how long it takes to find the symbol main.main,
// which is typically the last func symbol(by experiment).
func BenchmarkGetMain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var main_init func()
		_ = GetFunc(&main_init, "main.main")
	}
}

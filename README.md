# go-forceexport

go-forceexport is a golang package that allows access to any module-level
function, even ones that are not exported. You give it the string name of a
function , like `"time.now"`, and gives you a function value that calls that
function. More generally, it can be used to achieve something like reflection on
top-level functions, whereas the `reflect` package only lets you access methods
by name.

As you might expect, this library is **unsafe** and **fragile** and probably
shouldn't be used in production. See "Use cases and pitfalls" below.

It has only been tested on Mac OS X with Go 1.14/1.15/1.16. If you find that it works or
breaks on other platforms, feel free to submit a pull request with a fix and/or
an update to this paragraph.

## Installation

`$ go get github.com/AlaxLee/go-forceexport`

## Usage

Here's how you can grab the `time.now` function, defined as
`func now() (sec int64, nsec int32)`

```go
var timeNow func() (int64, int32)
err := forceexport.GetFunc(&timeNow, "time.now")
if err != nil {
    // Handle errors if you care about name possibly being invalid.
}
// Calls the actual time.now function.
sec, nsec := timeNow()
```

The string you give should be the fully-qualified name. For example, here's
`GetFunc` getting itself.

```go
var getFunc func(interface{}, string) error
GetFunc(&getFunc, "github.com/AlaxLee/go-forceexport.GetFunc")
```

**NOTICE:** Sometimes `GetFunc` could not find the wanted function. There are usually two reasons:

1. The function is optimized(inline). For example，to get the `time.(*Time).unixSec` function.
```go
	// unixSec returns the time's seconds since Jan 1 1970 (Unix time).
	// func (t *Time) unixSec() int64 { return t.sec() + internalToUnix }
	var unixSec func(time *time.Time) int64
	err := forceexport.GetFunc(&unixSec, "time.(*Time).unixSec")
	if err != nil {
		// Handle errors if you care about name possibly being invalid.
		panic(err)
	}
	usec := unixSec(&time.Time{})
	fmt.Println(usec)
	fmt.Println(time.Time{}.Unix())   // time.Time.Unix is same as time.(*Time).unixSec
```
We will receive an error "Invalid function name: time.(*Time).unixSec"
```text
AlaxdeMacBook-Pro:example alax$ go run main.go 
panic: Invalid function name: time.(*Time).unixSec
```
But if we don’t use optimization with flag "-l", we will get it successfully.
```text
AlaxdeMacBook-Pro:example alax$ go run -gcflags "all=-l"  main.go 
-62135596800
-62135596800
```
We can use "go build --gcflags=-m" to detect if optimized.
```text
AlaxdeMacBook-Pro:example alax$ cd $GOROOT/src/time
AlaxdeMacBook-Pro:time alax$ go build --gcflags=-m  2>&1 |grep -i inline|grep -i unixSec
./time.go:176:6: can inline (*Time).unixSec
```

2. The function is not used. For example，to get the `go/types.(*Checker).representable` function.
```go
//must be kept in sync with operand in src/go/types/operand.go
	type operandMode byte
	type builtinId int
	type operand struct {
		mode operandMode
		expr ast.Expr
		typ  types.Type
		val  constant.Value
		id   builtinId
	}
	//must same as method (*Checker).representable in src/go/types/expr.go
	var _representable func(checker *types.Checker, x *operand, typ *types.Basic)
	// 将 _representable 映射为 go/types.(*Checker).representable
	err := forceexport.GetFunc(&_representable, "go/types.(*Checker).representable")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("OK")
	}
```

We will receive an error "Invalid function name: go/types.(*Checker).representable".
And it isn't because of optimized.
```text
AlaxdeMacBook-Pro:example alax$ go run main.go 
panic: Invalid function name: go/types.(*Checker).representable
AlaxdeMacBook-Pro:example alax$ go run -gcflags "all=-l" main.go
panic: Invalid function name: go/types.(*Checker).representable
AlaxdeMacBook-Pro:example alax$ cd $GOROOT/src/go/types
AlaxdeMacBook-Pro:types alax$ grep ' representable(' *.go
expr.go:func (check *Checker) representable(x *operand, typ *Basic) {
AlaxdeMacBook-Pro:types alax$ go build --gcflags=-m  2>&1 | grep -i representable
AlaxdeMacBook-Pro:types alax$ 
```
Before `GetFunc`, we use it.
```go
	var packageName = "haha"
	var code = `
package haha
func main() {}
`
	var err error
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, packageName+".go", code, 0)
	if err != nil {
		log.Panicf("parse code failed: %s", err)
	}
	c := new(types.Config)
	c.Error = func(err error) {}
	pkg := types.NewPackage(packageName, "")
	checker := types.NewChecker(c, fset, pkg, nil)
	err = checker.Files([]*ast.File{file}) // (* Checker).Files use the wanted function
	if err != nil {
		log.Panicf("check file failed: %s", err)
	}

	//must be kept in sync with operand in src/go/types/operand.go
	type operandMode byte
	type builtinId int
	type operand struct {
		mode operandMode
		expr ast.Expr
		typ  types.Type
		val  constant.Value
		id   builtinId
	}
	//must same as method (*Checker).representable in src/go/types/expr.go
	var _representable func(checker *types.Checker, x *operand, typ *types.Basic)
	// 将 _representable 映射为 go/types.(*Checker).representable
	err = forceexport.GetFunc(&_representable, "go/types.(*Checker).representable")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("OK")
	}
```
And we got it.
```text
AlaxdeMacBook-Pro:example alax$ go run main.go 
OK
```
Maybe the call link is as follows.
```text
func (check *Checker) Files(files []*ast.File) error
-> func (check *Checker) checkFiles(files []*ast.File) (err error)
-> func (check *Checker) packageObjects()
-> func (check *Checker) objDecl(obj Object, def *Named)
-> func (check *Checker) funcDecl(obj *Func, decl *declInfo)
-> func (check *Checker) funcBody(decl *declInfo, name string, sig *Signature, body *ast.BlockStmt, iota constant.Value)
-> func (check *Checker) stmtList(ctxt stmtContext, list []ast.Stmt)
-> func (check *Checker) stmt(ctxt stmtContext, s ast.Stmt)
-> func (check *Checker) rawExpr(x *operand, e ast.Expr, hint Type) exprKind
-> func (check *Checker) exprInternal(x *operand, e ast.Expr, hint Type) exprKind
-> func (check *Checker) binary(x *operand, e *ast.BinaryExpr, lhs, rhs ast.Expr, op token.Token, opPos token.Pos)
-> func (check *Checker) shift(x, y *operand, e *ast.BinaryExpr, op token.Token)
-> func (check *Checker) representable(x *operand, typ *Basic)
```

## Use cases and pitfalls

This library is most useful for development and hack projects. For example, you
might use it to track down why the standard library isn't behaving as you
expect, or you might use it to try out a standard library function to see if it
works, then later factor the code to be less fragile. You could also try using
it in production; just make sure you're aware of the risks.

There are lots of things to watch out for and ways to shoot yourself in
the foot:
* If you define the wrong function type, you'll get a function with undefined
  behavior that will likely cause a runtime panic. The library makes no attempt
  to warn you in this case.
* Calling unexported functions is inherently fragile because the function won't
  have any stability guarantees.
* The implementation relies on the details of internal Go data structures, so
  later versions of Go might break this library.
* Since the compiler doesn't expect unexported symbols to be used, it might not
  create them at all, for example due to inlining or dead code analysis. This
  means that functions may not show up like you expect, and new versions of the
  compiler may cause functions to suddenly disappear.
* If the function you want to use relies on unexported types, you won't be able
  to trivially use it. However, you can sometimes work around this by defining
  equivalent copies of those types that you can use, but that approach has its
  own set of dangers.

## How it works

The [code](/forceexport.go) is pretty short, so you could just read it, but
here's a friendlier explanation:

The code uses the `go:linkname` compiler directive to get access to the
`runtime.firstmoduledata` symbol, which is an internal data structure created by
the linker that's used by functions like `runtime.FuncForPC`. (Using
`go:linkname` is an alternate way to access unexported functions/values, but it
has other gotchas and can't be used dynamically.)

Similar to the implementation of `runtime.FuncForPC`, the code walks the
function definitions until it finds one with a matching name, then gets its code
pointer.

From there, it creates a function object from the code pointer by calling
`reflect.MakeFunc` and using `unsafe.Pointer` to swap out the function object's
code pointer with the desired one.

Needless to say, it's a scary hack, but it seems to work!

## Thanks

https://github.com/alangpierce/go-forceexport

https://github.com/linux4life798/go-forceexport

https://github.com/zhuzhengyang/go-forceexport

## License

MIT
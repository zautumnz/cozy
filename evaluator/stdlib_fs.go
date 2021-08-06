package evaluator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/zacanger/cozy/object"
)

// array = fs.glob("/etc/*.conf")
func fsGlob(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	pattern := args[0].(*object.String).Value

	entries, err := filepath.Glob(pattern)
	if err != nil {
		return &object.Error{Message: err.Error()}
	}

	// Create an array to hold the results and populate it
	l := len(entries)
	result := make([]object.Object, l)
	for i, txt := range entries {
		result[i] = &object.String{Value: txt}
	}
	return &object.Array{Elements: result}
}

// Change a mode of a file - note the second argument is a string
// to emphasise octal.
func chmodFun(args ...object.Object) object.Object {
	if len(args) != 2 {
		return NewError("wrong number of arguments. got=%d, want=2",
			len(args))
	}

	path := args[0].Inspect()
	mode := ""

	switch args[1].(type) {
	case *object.String:
		mode = args[1].(*object.String).Value
	default:
		return NewError("Second argument must be string, got %v", args[1])
	}

	// convert from octal -> decimal
	result, err := strconv.ParseInt(mode, 8, 64)
	if err != nil {
		return &object.Boolean{Value: false}
	}

	// Change the mode.
	err = os.Chmod(path, os.FileMode(result))
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}
}

// mkdir
func mkdirFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	if args[0].Type() != object.STRING_OBJ {
		return NewError("argument to `mkdir` must be STRING, got %s",
			args[0].Type())
	}

	path := args[0].(*object.String).Value

	// Can't fail?
	mode, err := strconv.ParseInt("755", 8, 64)
	if err != nil {
		return &object.Boolean{Value: false}
	}

	err = os.MkdirAll(path, os.FileMode(mode))
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}

}

// Open a file
func openFun(args ...object.Object) object.Object {
	path := ""
	mode := "r"

	// We need at least one arg
	if len(args) < 1 {
		return NewError("wrong number of arguments. got=%d, want=1+",
			len(args))
	}

	// Get the filename
	switch args[0].(type) {
	case *object.String:
		path = args[0].(*object.String).Value
	default:
		return NewError("argument to `file` not supported, got=%s",
			args[0].Type())

	}

	// Get the mode (optiona)
	if len(args) > 1 {
		switch args[1].(type) {
		case *object.String:
			mode = args[1].(*object.String).Value
		default:
			return NewError("argument to `file` not supported, got=%s",
				args[0].Type())

		}
	}

	// Create the object
	file := &object.File{Filename: path}
	file.Open(mode)
	return (file)
}

// Get file info.
func statFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}
	path := args[0].Inspect()
	info, err := os.Stat(path)

	res := make(map[object.HashKey]object.HashPair)
	if err != nil {
		// Empty hash as we've not yet set anything
		return &object.Hash{Pairs: res}
	}

	// Populate the hash

	// size -> int
	sizeData := &object.Integer{Value: info.Size()}
	sizeKey := &object.String{Value: "size"}
	sizeHash := object.HashPair{Key: sizeKey, Value: sizeData}
	res[sizeKey.HashKey()] = sizeHash

	// mod-time -> int
	mtimeData := &object.Integer{Value: info.ModTime().Unix()}
	mtimeKey := &object.String{Value: "mtime"}
	mtimeHash := object.HashPair{Key: mtimeKey, Value: mtimeData}
	res[mtimeKey.HashKey()] = mtimeHash

	// Perm -> string
	permData := &object.String{Value: info.Mode().String()}
	permKey := &object.String{Value: "perm"}
	permHash := object.HashPair{Key: permKey, Value: permData}
	res[permKey.HashKey()] = permHash

	// Mode -> string  (because we want to emphasise the octal nature)
	m := fmt.Sprintf("%04o", info.Mode().Perm())
	modeData := &object.String{Value: m}
	modeKey := &object.String{Value: "mode"}
	modeHash := object.HashPair{Key: modeKey, Value: modeData}
	res[modeKey.HashKey()] = modeHash

	typeStr := "unknown"
	if info.Mode().IsDir() {
		typeStr = "directory"
	}
	if info.Mode().IsRegular() {
		typeStr = "file"
	}

	// type: string
	typeData := &object.String{Value: typeStr}
	typeKey := &object.String{Value: "type"}
	typeHash := object.HashPair{Key: typeKey, Value: typeData}
	res[typeKey.HashKey()] = typeHash

	return &object.Hash{Pairs: res}

}

// Remove a file/directory.
func unlinkFun(args ...object.Object) object.Object {
	if len(args) != 1 {
		return NewError("wrong number of arguments. got=%d, want=1",
			len(args))
	}

	path := args[0].Inspect()

	err := os.Remove(path)
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}
}

func mvFun(args ...object.Object) object.Object {
	var from string
	var to string
	switch a := args[0].(type) {
	case *object.String:
		from = a.Value
	default:
		return NewError("mv expected string arg!")
	}
	switch a := args[1].(type) {
	case *object.String:
		to = a.Value
	default:
		return NewError("mv expected string arg!")
	}

	e := os.Rename(from, to)
	if e != nil {
		return NewError("error moving file %s", e.Error())
	}

	return NULL
}

func cpFun(args ...object.Object) object.Object {
	var src string
	var dst string
	switch a := args[0].(type) {
	case *object.String:
		src = a.Value
	default:
		return NewError("mv expected string arg!")
	}
	switch a := args[1].(type) {
	case *object.String:
		dst = a.Value
	default:
		return NewError("mv expected string arg!")
	}

	sfi, err := os.Stat(src)
	if err != nil {
		return NewError("fs.cp source does not exist!")
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return NewError("fs.cp expected regular file!")
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return NewError("error copying file %s", err.Error())
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return NewError("non-regular destination file")
		}
		if os.SameFile(sfi, dfi) {
			return NewError("copying to same file")
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return NewError("error copying file %s", err.Error())
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return NewError("error copying file %s", err.Error())
	}

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return NewError("error copying file %s", err.Error())
	}
	err = out.Sync()

	if err != nil {
		return NewError("error copying file %s", err.Error())
	}

	return NULL
}

func init() {
	RegisterBuiltin("fs.glob",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (fsGlob(args...))
		})
	RegisterBuiltin("fs.chmod",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (chmodFun(args...))
		})
	RegisterBuiltin("fs.mkdir",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (mkdirFun(args...))
		})
	RegisterBuiltin("fs.open",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (openFun(args...))
		})
	RegisterBuiltin("fs.stat",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (statFun(args...))
		})
	RegisterBuiltin("fs.unlink",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (unlinkFun(args...))
		})
	RegisterBuiltin("fs.mv",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (mvFun(args...))
		})
	RegisterBuiltin("fs.cp",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (cpFun(args...))
		})
}

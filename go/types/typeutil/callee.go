// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package typeutil

import (
	"go/ast"
	"go/types"
	_ "unsafe" // for linkname
)

// Callee returns the named target of a function call, if any:
// a function, method, builtin, or variable.
//
// Functions and methods may potentially have type parameters.
//
// Note: for calls of instantiated functions and methods, Callee returns
// the corresponding generic function or method on the generic type.
func Callee(info *types.Info, call *ast.CallExpr) types.Object {
	obj := used(info, call.Fun)
	if obj == nil {
		return nil
	}
	if _, ok := obj.(*types.TypeName); ok {
		return nil
	}
	return obj
}

// StaticCallee returns the target (function or method) of a static function
// call, if any. It returns nil for calls to builtins.
//
// Note: for calls of instantiated functions and methods, StaticCallee returns
// the corresponding generic function or method on the generic type.
func StaticCallee(info *types.Info, call *ast.CallExpr) *types.Func {
	obj := used(info, call.Fun)
	fn, _ := obj.(*types.Func)
	if fn == nil || interfaceMethod(fn) {
		return nil
	}
	return fn
}

// used is the implementation of [internal/typesinternal.Used].
// It returns the object associated with e.
// See typesinternal.Used for a fuller description.
// This function should live in typesinternal, but cannot because it would
// create an import cycle.
//
//go:linkname used
func used(info *types.Info, e ast.Expr) types.Object {
	if info.Types == nil || info.Uses == nil || info.Selections == nil {
		panic("one of info.Types, info.Uses or info.Selections is nil; all must be populated")
	}
	// Look through type instantiation if necessary.
	switch d := ast.Unparen(e).(type) {
	case *ast.IndexExpr:
		if info.Types[d.Index].IsType() {
			e = d.X
		}
	case *ast.IndexListExpr:
		e = d.X
	}

	var obj types.Object
	switch e := ast.Unparen(e).(type) {
	case *ast.Ident:
		obj = info.Uses[e] // type, var, builtin, or declared func
	case *ast.SelectorExpr:
		if sel, ok := info.Selections[e]; ok {
			obj = sel.Obj() // method or field
		} else {
			obj = info.Uses[e.Sel] // qualified identifier?
		}
	}
	return obj
}

// interfaceMethod reports whether its argument is a method of an interface.
// This function should live in typesinternal, but cannot because it would create an import cycle.
//
//go:linkname interfaceMethod
func interfaceMethod(f *types.Func) bool {
	recv := f.Signature().Recv()
	return recv != nil && types.IsInterface(recv.Type())
}

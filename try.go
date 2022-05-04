// Copyright 2022, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// Package try emulates aspects of the ill-fated "try" proposal using generics.
// See https://golang.org/issue/32437 for inspiration.
//
// Example usage:
//
//	func Fizz(...) (..., err error) {
//		defer try.HandleF(&err, func() {
//			if err == io.EOF {
//				err = io.ErrUnexpectedEOF
//			}
//		})
//		... := try.E2(Buzz(...))
//		return ..., nil
//	}
//
// This package is not intended for production critical code.
// Quick and easy error handling can occlude critical error handling logic.
// Rather, it is intended for short Go programs and unit tests where
// development speed is a greater priority than reliability.
// Since the E functions panic if an error is encountered, recovering is optional.
//
//
// Code before try:
//
//	func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) error {
//		switch t, err := d.ReadToken(); {
//		case err != nil:
//			return err
//		case t.Kind() != '[':
//			return fmt.Errorf("got %v, expecting array start", t.Kind())
//		}
//
//		if err := uo.UnmarshalNext(d, &a.Scalar); err != nil {
//			return err
//		}
//		if err := uo.UnmarshalNext(d, &a.Slice); err != nil {
//			return err
//		}
//		if err := uo.UnmarshalNext(d, &a.Map); err != nil {
//			return err
//		}
//
//		switch t, err := d.ReadToken(); {
//		case err != nil:
//			return err
//		case t.Kind() != ']':
//			return fmt.Errorf("got %v, expecting array start", t.Kind())
//		}
//		return nil
//	}
//
// Code after try:
//
//	func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) (err error) {
//		defer try.Handle(&err)
//		if t := try.E1(d.ReadToken()); t.Kind() != '[' {
//			return fmt.Errorf("found %v, expecting array start", t.Kind())
//		}
//		try.E(uo.UnmarshalNext(d, &a.Scalar))
//		try.E(uo.UnmarshalNext(d, &a.Slice))
//		try.E(uo.UnmarshalNext(d, &a.Map))
//		if t := try.E1(d.ReadToken()); t.Kind() != ']' {
//			return fmt.Errorf("found %v, expecting array start", t.Kind())
//		}
//		return nil
//	}
//
package try

import (
	"log"
	"runtime"
	"testing"
)

// wrapError wraps an error to ensure that we only recover from errors
// panicked by this package.
type wrapError struct{ error }

// Unwrap primarily exists for testing purposes.
func (e wrapError) Unwrap() error { return e.error }

// Recover recovers a previously panicked error and stores it into err.
// If it successfully recovers an error, and fn is non-nil,
// it calls fn with the runtime frame in which the error occurred.
//
// Recover is a general purpose API.
// Most use cases will be better served by Handle, HandleF, TB, or Fatal.
func Recover(errptr *error, fn func(runtime.Frame)) {
	r(recover(), errptr, fn)
}

// r implements recover.
// It is a separate function from Recover to keep stack counts consistent.
func r(recovered any, err *error, fn func(runtime.Frame)) {
	switch ex := recovered.(type) {
	case nil:
		return
	case wrapError:
		*err = ex.error
		if fn != nil {
			pc := make([]uintptr, 1)
			// 5: runtime.Callers, r, Recover/Handle/etc, the function that called defer Recover, the actual panic.
			n := runtime.Callers(5, pc)
			pc = pc[:n]
			frames := runtime.CallersFrames(pc)
			frame, _ := frames.Next()
			fn(frame)
		}
	default:
		panic(ex)
	}
}

// Handle recovers a previously panicked error and stores it into err.
func Handle(errptr *error) {
	r(recover(), errptr, nil)
}

// HandleF recovers a previously panicked error and stores it into err.
// If it successfully recovers an error, it calls fn.
func HandleF(errptr *error, fn func()) {
	r(recover(), errptr, func(runtime.Frame) { fn() })
}

// TB recovers any panicked errors from this package and calls tb.Fatalf.
// It is useful for simple tests and benchmarks:
//
// func TestFoo(t *testing.T) {
//   defer try.TB(t)
//   // use try.E throughout your test
// }
func TB(tb testing.TB) {
	var err error
	r(recover(), &err, func(frame runtime.Frame) {
		tb.Fatalf("%s:%d %v", frame.File, frame.Line, err)
	})
}

// Fatal recovers any panicked errors from this package and calls log.Fatalf.
// It is useful in quick-and-dirty scripts:
//
// func main() {
//   defer try.Fatal()
//   // use try.E throughout your program
// }
func Fatal() {
	var err error
	r(recover(), &err, func(frame runtime.Frame) {
		log.Fatalf("%s:%d %v", frame.File, frame.Line, err)
	})
}

// E panics if err is non-nil.
func E(err error) {
	if err != nil {
		panic(wrapError{err})
	}
}

// E1 returns a as is.
// It panics if err is non-nil.
func E1[A any](a A, err error) A {
	E(err)
	return a
}

// E2 returns a and b as is.
// It panics if err is non-nil.
func E2[A, B any](a A, b B, err error) (A, B) {
	E(err)
	return a, b
}

// E3 returns a, b, and c as is.
// It panics if err is non-nil.
func E3[A, B, C any](a A, b B, c C, err error) (A, B, C) {
	E(err)
	return a, b, c
}

// E4 returns a, b, c, and d as is.
// It panics if err is non-nil.
func E4[A, B, C, D any](a A, b B, c C, d D, err error) (A, B, C, D) {
	E(err)
	return a, b, c, d
}

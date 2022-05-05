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
//
// Quick tour of the API
//
//
// The E family of functions all remove a final error return, panicking if non-nil.
//
//
// Handle allows easy assignment of that error to a return error value.
//
//  func f() (err error) {
//  	defer try.Handle(&err)
//
//
// HandleF is like Handle, but it calls a function after any such assignment.
//
//  func f() (err error) {
//		defer try.HandleF(&err, func() {
//			if err == io.EOF {
//				err = io.ErrUnexpectedEOF
//			}
//		})
//
// F wraps an error with file and line formation and calls a function on error.
// It plays nicely with testing.TB and log.Fatal.
//
//  func TestFoo(t *testing.T) {
//  	defer try.F(t.Fatal)
//
//  func main() {
//  	defer try.F(log.Fatal)
//
//
// Recover is like F, but it supports more complicated error handling
// by passing the error and runtime frame directly to a function.
//
//  func f() {
//  	defer try.Recover(func(err error, frame runtime.Frame) {
//  		// do something useful with err and frame
// 		})
//
//
package try

import (
	"fmt"
	"runtime"
)

// wrapError wraps an error to ensure that we only recover from errors
// panicked by this package.
type wrapError struct {
	error
	frame runtime.Frame
}

// Unwrap primarily exists for testing purposes.
func (e wrapError) Unwrap() error {
	return e.error
}

func r(recovered any, fn func(err error, frame runtime.Frame)) {
	switch ex := recovered.(type) {
	case nil:
	case wrapError:
		fn(ex.error, ex.frame)
	default:
		panic(ex)
	}
}

// Recover recovers an error previously panicked with an E function.
// If it recovers an error, it calls fn with the error and the runtime frame in which it occurred.
func Recover(fn func(err error, frame runtime.Frame)) {
	r(recover(), fn)
}

// Handle recovers an error previously panicked with an E function and stores it into errptr.
func Handle(errptr *error) {
	r(recover(), func(err error, _ runtime.Frame) {
		*errptr = err
	})
}

// HandleF recovers an error previously panicked with an E function and stores it into errptr.
// If it recovers an error, it calls fn.
func HandleF(errptr *error, fn func()) {
	r(recover(), func(err error, _ runtime.Frame) {
		*errptr = err
		if err != nil {
			fn()
		}
	})
}

// F recovers an error previously panicked with an E function, wraps it, and passes it to fn.
// The wrapping includes the file and line of the runtime frame in which it occurred.
// F pairs well with testing.TB.Fatal and log.Fatal.
func F(fn func(...any)) {
	r(recover(), func(err error, frame runtime.Frame) {
		fn(fmt.Errorf("%s:%d: %w", frame.File, frame.Line, err))
	})
}

func e(err error) {
	if err != nil {
		pc := make([]uintptr, 1)
		// 3: runtime.Callers, e, E
		n := runtime.Callers(3, pc)
		pc = pc[:n]
		frames := runtime.CallersFrames(pc)
		frame, _ := frames.Next()
		panic(wrapError{error: err, frame: frame})
	}
}

// E panics if err is non-nil.
func E(err error) {
	e(err)
}

// E1 returns a as is.
// It panics if err is non-nil.
func E1[A any](a A, err error) A {
	e(err)
	return a
}

// E2 returns a and b as is.
// It panics if err is non-nil.
func E2[A, B any](a A, b B, err error) (A, B) {
	e(err)
	return a, b
}

// E3 returns a, b, and c as is.
// It panics if err is non-nil.
func E3[A, B, C any](a A, b B, c C, err error) (A, B, C) {
	e(err)
	return a, b, c
}

// E4 returns a, b, c, and d as is.
// It panics if err is non-nil.
func E4[A, B, C, D any](a A, b B, c C, d D, err error) (A, B, C, D) {
	e(err)
	return a, b, c, d
}

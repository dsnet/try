// Copyright 2022, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

// Package try emulates aspects of the ill-fated "try" proposal using generics.
// See https://golang.org/issue/32437 for inspiration.
//
// Example usage:
//
//	func Fizz(...) (..., err error) {
//		defer try.Catch(&err, func() {
//			if err == io.EOF {
//				err = io.ErrUnexpectedEOF
//			}
//		})
//		... := try.E2(Buzz(...))
//		return ..., nil
//	}
//
// This package is not intended for production critical code as quick and easy
// error handling can occlude critical error handling logic.
// Rather, it is intended for short Go programs and unit tests where
// development speed is a greater priority than reliability.
// Since the E functions panic if an error is encountered,
// calling Catch in a deferred function is optional.
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
//		defer try.Catch(&err)
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

// wrapError wraps an error to ensure that we only recover from errors
// panicked by this package.
type wrapError struct{ error }

// Unwrap primarily exists for testing purposes.
func (e wrapError) Unwrap() error { return e.error }

// Catch catches a previously panicked error and stores it into err.
// If it successfully catches an error, it calls any provided handlers.
func Catch(err *error, handlers ...func()) {
	switch ex := recover().(type) {
	case nil:
		return
	case wrapError:
		*err = ex.error
		for _, handler := range handlers {
			handler()
		}
	default:
		panic(ex)
	}
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

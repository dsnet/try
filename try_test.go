// Copyright 2022, Joe Tsai. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package try_test

import (
	"errors"
	"io"
	"log"
	"runtime"
	"strings"
	"testing"

	"github.com/dsnet/try"
)

func Test(t *testing.T) {
	tests := []struct {
		name      string
		run       func(*testing.T) error
		wantError error
		wantPanic error
	}{{
		name: "NoRecover/Success",
		run: func(t *testing.T) error {
			a, b, c := try.E3(success())
			if a != 1 && b != "success" && c != true {
				t.Errorf("success() = (%v, %v, %v), want (1, success, true)", a, b, c)
			}
			return nil
		},
	}, {
		name: "NoRecover/Failure",
		run: func(t *testing.T) error {
			a, b, c := try.E3(failure())
			t.Errorf("failure() = (%v, %v, %v), want panic", a, b, c)
			return nil
		},
		wantPanic: io.EOF,
	}, {
		name: "Recover/Success",
		run: func(t *testing.T) (err error) {
			defer try.Handle(&err)
			a, b, c := try.E3(success())
			if a != 1 && b != "success" && c != true {
				t.Errorf("success() = (%v, %v, %v), want (1, success, true)", a, b, c)
			}
			return nil
		},
	}, {
		name: "Recover/Failure",
		run: func(t *testing.T) (err error) {
			defer try.Handle(&err)
			a, b, c := try.E3(failure())
			t.Errorf("failure() = (%v, %v, %v), want panic", a, b, c)
			return nil
		},
		wantError: io.EOF,
	}, {
		name: "Recover/Failure/Ignored",
		run: func(t *testing.T) (err error) {
			defer try.HandleF(&err, func() {
				if err == io.EOF {
					err = nil
				}
			})
			a, b, c := try.E3(failure())
			t.Errorf("failure() = (%v, %v, %v), want panic", a, b, c)
			return nil
		},
	}, {
		name: "Recover/Failure/Replaced",
		run: func(t *testing.T) (err error) {
			defer try.HandleF(&err, func() {
				if err == io.EOF {
					err = io.ErrUnexpectedEOF
				}
			})
			a, b, c := try.E3(failure())
			t.Errorf("failure() = (%v, %v, %v), want panic", a, b, c)
			return nil
		},
		wantError: io.ErrUnexpectedEOF,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotError error
			var gotPanic error
			func() {
				defer func() {
					r := recover()
					if r == nil {
						return
					}
					var ok bool
					gotPanic, ok = r.(error)
					if !ok {
						t.Errorf("recovered non-error %T", r)
					}
				}()
				gotError = tt.run(t)
			}()
			switch {
			case !errors.Is(gotError, tt.wantError):
				t.Errorf("returned error: got %v, want %v", gotError, tt.wantError)
			case !errors.Is(gotPanic, tt.wantPanic):
				t.Errorf("panicked error: got %v, want %v", gotPanic, tt.wantPanic)
			}
		})
	}
}

func TestFrame(t *testing.T) {
	defer try.Recover(func(err error, frame runtime.Frame) {
		if frame.File != "x.go" {
			t.Errorf("want File=x.go, got %q", frame.File)
		}
		if frame.Line != 4 {
			t.Errorf("want Line=4, got %d", frame.Line)
		}
	})
//line x.go:4
	try.E(errors.New("crash and burn"))
}

func TestF(t *testing.T) {
	buf := new(strings.Builder)
	logger := log.New(buf, "", 0)
	defer func() {
		const want = "y.go:10 EOF\n"
		if got := buf.String(); got != want {
			t.Errorf("want %q, got %q",want, got)
		}
	}()
	defer try.F(logger.Print)
//line y.go:10
	try.E(io.EOF)
}

func success() (a int, b string, c bool, err error) {
	return +1, "success", true, nil
}

func failure() (a int, b string, c bool, err error) {
	return -1, "failure", false, io.EOF
}

func BenchmarkSuccess(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() (err error) {
			defer try.Handle(&err)
			try.E3(success())
			return nil
		}()
	}
}

func BenchmarkFailure(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() (err error) {
			defer try.Handle(&err)
			try.E3(failure())
			return nil
		}()
	}
}

# Try: Simplified Error Handling in Go

[![GoDev](https://img.shields.io/static/v1?label=godev&message=reference&color=00add8)][godev]
[![Build Status](https://github.com/dsnet/try/actions/workflows/test.yml/badge.svg?branch=master)][actions]

This module reduces the syntactic cost of error handling in Go.

Example usage in a main program:

```go
func main() {
    defer try.F(log.Fatal)
    b := try.E1(os.ReadFile(...))
    var v any
    try.E(json.Unmarshal(b, &v))
    ...
}
```

Example usage in a unit test:

```go
func Test(t *testing.T) {
    defer try.F(t.Fatal)
    db := try.E1(setdb.Open(...))
    defer db.Close()
    ...
    try.E(db.Commit())
}
```

Code before `try`:

```go
func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) error {
    switch t, err := d.ReadToken(); {
    case err != nil:
        return err
    case t.Kind() != '[':
        return fmt.Errorf("got %v, expecting array start", t.Kind())
    }

    if err := uo.UnmarshalNext(d, &a.Scalar); err != nil {
        return err
    }
    if err := uo.UnmarshalNext(d, &a.Slice); err != nil {
        return err
    }
    if err := uo.UnmarshalNext(d, &a.Map); err != nil {
        return err
    }

    switch t, err := d.ReadToken(); {
    case err != nil:
        return err
    case t.Kind() != ']':
        return fmt.Errorf("got %v, expecting array end", t.Kind())
    }
    return nil
}
```

Code after `try`:

```go
func (a *MixedArray) UnmarshalNext(uo json.UnmarshalOptions, d *json.Decoder) (err error) {
    defer try.Handle(&err)
    if t := try.E1(d.ReadToken()); t.Kind() != '[' {
        return fmt.Errorf("found %v, expecting array start", t.Kind())
    }
    try.E(uo.UnmarshalNext(d, &a.Scalar))
    try.E(uo.UnmarshalNext(d, &a.Slice))
    try.E(uo.UnmarshalNext(d, &a.Map))
    if t := try.E1(d.ReadToken()); t.Kind() != ']' {
        return fmt.Errorf("found %v, expecting array end", t.Kind())
    }
    return nil
}
```

See the [documentation][godev] for more information.

[godev]: https://pkg.go.dev/github.com/dsnet/try
[actions]: https://github.com/dsnet/try/actions

## Install

```
go get -u github.com/dsnet/try
```

## License

BSD - See [LICENSE][license] file

[license]: https://github.com/dsnet/try/blob/master/LICENSE.md
// Copyright 2018 The Wire Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wire

import (
	"errors"
	"go/token"
	"testing"
)

func TestNotePosition(t *testing.T) {
	t.Run("nil returns nil", func(t *testing.T) {
		got := notePosition(token.Position{Filename: "foo.go", Line: 1}, nil)
		if got != nil {
			t.Errorf("notePosition(pos, nil) = %v; want nil", got)
		}
	})

	t.Run("plain error gets wrapped", func(t *testing.T) {
		pos := token.Position{Filename: "foo.go", Line: 10, Column: 5}
		err := errors.New("something failed")
		got := notePosition(pos, err)
		we, ok := got.(*wireErr)
		if !ok {
			t.Fatalf("notePosition returned %T; want *wireErr", got)
		}
		if we.position != pos {
			t.Errorf("position = %v; want %v", we.position, pos)
		}
		if we.error != err {
			t.Errorf("inner error = %v; want %v", we.error, err)
		}
	})

	t.Run("wireErr is not re-wrapped", func(t *testing.T) {
		pos1 := token.Position{Filename: "inner.go", Line: 5}
		inner := &wireErr{error: errors.New("inner"), position: pos1}
		pos2 := token.Position{Filename: "outer.go", Line: 20}
		got := notePosition(pos2, inner)
		if got != inner {
			t.Errorf("notePosition should return the same *wireErr; got %v, want %v", got, inner)
		}
		// Position should be preserved from inner (deeper call has more precise info).
		we := got.(*wireErr)
		if we.position != pos1 {
			t.Errorf("position = %v; want %v (original)", we.position, pos1)
		}
	})
}

func TestNotePositionAll(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		got := notePositionAll(token.Position{}, nil)
		if got != nil {
			t.Errorf("notePositionAll(pos, nil) = %v; want nil", got)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		pos := token.Position{Filename: "test.go", Line: 1}
		errs := []error{
			errors.New("error 1"),
			errors.New("error 2"),
		}
		got := notePositionAll(pos, errs)
		if len(got) != 2 {
			t.Fatalf("got %d errors; want 2", len(got))
		}
		for i, err := range got {
			we, ok := err.(*wireErr)
			if !ok {
				t.Errorf("error[%d] is %T; want *wireErr", i, err)
				continue
			}
			if we.position != pos {
				t.Errorf("error[%d].position = %v; want %v", i, we.position, pos)
			}
		}
	})
}

func TestWireErrError(t *testing.T) {
	t.Run("invalid position", func(t *testing.T) {
		we := &wireErr{error: errors.New("boom"), position: token.Position{}}
		got := we.Error()
		if got != "boom" {
			t.Errorf("Error() = %q; want %q", got, "boom")
		}
	})

	t.Run("valid position", func(t *testing.T) {
		we := &wireErr{
			error:    errors.New("something broke"),
			position: token.Position{Filename: "foo.go", Line: 42, Column: 7},
		}
		got := we.Error()
		want := "foo.go:42:7: something broke"
		if got != want {
			t.Errorf("Error() = %q; want %q", got, want)
		}
	})
}

func TestErrorCollector(t *testing.T) {
	t.Run("nil errors are filtered", func(t *testing.T) {
		ec := new(errorCollector)
		ec.add(nil, nil, nil)
		if len(ec.errors) != 0 {
			t.Errorf("got %d errors; want 0", len(ec.errors))
		}
	})

	t.Run("single error added", func(t *testing.T) {
		ec := new(errorCollector)
		err := errors.New("test")
		ec.add(err)
		if len(ec.errors) != 1 {
			t.Fatalf("got %d errors; want 1", len(ec.errors))
		}
		if ec.errors[0] != err {
			t.Errorf("got error %v; want %v", ec.errors[0], err)
		}
	})

	t.Run("mixed nil and non-nil", func(t *testing.T) {
		ec := new(errorCollector)
		e1 := errors.New("one")
		e2 := errors.New("two")
		ec.add(nil, e1, nil, e2, nil)
		if len(ec.errors) != 2 {
			t.Fatalf("got %d errors; want 2", len(ec.errors))
		}
		if ec.errors[0] != e1 || ec.errors[1] != e2 {
			t.Errorf("got errors %v; want [%v, %v]", ec.errors, e1, e2)
		}
	})
}

func TestMapErrors(t *testing.T) {
	t.Run("empty slice returns nil", func(t *testing.T) {
		got := mapErrors(nil, func(e error) error { return e })
		if got != nil {
			t.Errorf("mapErrors(nil, f) = %v; want nil", got)
		}
	})

	t.Run("transform applied", func(t *testing.T) {
		errs := []error{errors.New("a"), errors.New("b")}
		got := mapErrors(errs, func(e error) error {
			return errors.New("wrapped: " + e.Error())
		})
		if len(got) != 2 {
			t.Fatalf("got %d errors; want 2", len(got))
		}
		if got[0].Error() != "wrapped: a" {
			t.Errorf("got[0] = %q; want %q", got[0].Error(), "wrapped: a")
		}
		if got[1].Error() != "wrapped: b" {
			t.Errorf("got[1] = %q; want %q", got[1].Error(), "wrapped: b")
		}
	})
}

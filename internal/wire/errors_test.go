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
	"strings"
	"testing"
)

func TestErrorCollector(t *testing.T) {
	t.Run("add errors", func(t *testing.T) {
		ec := &errorCollector{}

		// Add single error
		err1 := errors.New("error 1")
		ec.add(err1)

		if len(ec.errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(ec.errors))
		}

		// Add multiple errors
		err2 := errors.New("error 2")
		err3 := errors.New("error 3")
		ec.add(err2, err3)

		if len(ec.errors) != 3 {
			t.Errorf("expected 3 errors, got %d", len(ec.errors))
		}
	})

	t.Run("mapErrors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		errs := []error{err1, err2}

		// Map errors with a transformation function
		result := mapErrors(errs, func(err error) error {
			return errors.New("mapped: " + err.Error())
		})

		if len(result) != 2 {
			t.Errorf("expected 2 errors, got %d", len(result))
		}

		for _, err := range result {
			if !strings.HasPrefix(err.Error(), "mapped:") {
				t.Errorf("expected error to be mapped, got: %v", err)
			}
		}
	})
}

func TestNotePosition(t *testing.T) {
	t.Run("with valid position", func(t *testing.T) {
		pos := token.Position{
			Filename: "test.go",
			Line:     10,
			Column:   5,
		}
		baseErr := errors.New("test error")

		result := notePosition(pos, baseErr)

		errStr := result.Error()
		if !strings.Contains(errStr, "test.go") {
			t.Errorf("error should contain filename, got: %v", errStr)
		}
		if !strings.Contains(errStr, "test error") {
			t.Errorf("error should contain original message, got: %v", errStr)
		}
	})

	t.Run("with invalid position", func(t *testing.T) {
		pos := token.Position{}
		baseErr := errors.New("test error")

		result := notePosition(pos, baseErr)

		// notePosition wraps even with invalid position
		if _, ok := result.(*wireErr); !ok {
			t.Errorf("expected wireErr type, got: %T", result)
		}
		// The error message should be the same as base error for invalid position
		if result.Error() != baseErr.Error() {
			t.Errorf("error message = %v, want %v", result.Error(), baseErr.Error())
		}
	})
}

func TestNotePositionAll(t *testing.T) {
	t.Run("note position on multiple errors", func(t *testing.T) {
		pos := token.Position{
			Filename: "test.go",
			Line:     10,
			Column:   5,
		}

		err1 := errors.New("error 1")
		err2 := errors.New("error 2")

		results := notePositionAll(pos, []error{err1, err2})

		if len(results) != 2 {
			t.Errorf("expected 2 errors, got %d", len(results))
		}

		for _, err := range results {
			errStr := err.Error()
			if !strings.Contains(errStr, "test.go") {
				t.Errorf("error should contain filename, got: %v", errStr)
			}
		}
	})
}

func TestWireErr_Error(t *testing.T) {
	t.Run("error with valid position", func(t *testing.T) {
		baseErr := errors.New("test error")
		pos := token.Position{
			Filename: "test.go",
			Line:     10,
			Column:   5,
		}
		we := &wireErr{error: baseErr, position: pos}
		errStr := we.Error()

		if !strings.Contains(errStr, "test.go") {
			t.Errorf("wireErr.Error() should contain filename, got: %v", errStr)
		}
		if !strings.Contains(errStr, "test error") {
			t.Errorf("wireErr.Error() should contain error message, got: %v", errStr)
		}
	})

	t.Run("error with invalid position", func(t *testing.T) {
		baseErr := errors.New("test error")
		we := &wireErr{error: baseErr, position: token.Position{}}
		errStr := we.Error()

		if errStr != "test error" {
			t.Errorf("wireErr.Error() with invalid position should return original message, got: %v", errStr)
		}
	})
}

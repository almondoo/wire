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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateResult_Commit(t *testing.T) {
	tests := []struct {
		name       string
		content    []byte
		outputPath string
		wantErr    bool
	}{
		{
			name:       "empty content",
			content:    nil,
			outputPath: "",
			wantErr:    false,
		},
		{
			name:       "valid content",
			content:    []byte("package main\n\nfunc main() {}\n"),
			outputPath: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tmpDir, err := ioutil.TempDir("", "wire_test_commit")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			outputPath := tt.outputPath
			if tt.content != nil && outputPath == "" {
				outputPath = filepath.Join(tmpDir, "test_wire_gen.go")
			}

			gen := GenerateResult{
				PkgPath:    "test/pkg",
				OutputPath: outputPath,
				Content:    tt.content,
			}

			err = gen.Commit()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateResult.Commit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If we expected success and had content, verify the file was written
			if !tt.wantErr && tt.content != nil {
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Errorf("GenerateResult.Commit() did not create file at %s", outputPath)
				} else {
					// Verify content
					written, err := ioutil.ReadFile(outputPath)
					if err != nil {
						t.Fatalf("failed to read written file: %v", err)
					}
					if string(written) != string(tt.content) {
						t.Errorf("GenerateResult.Commit() wrote incorrect content.\nGot:  %s\nWant: %s", written, tt.content)
					}
				}
			}
		})
	}
}

func TestZeroValue(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected string
	}{
		{
			name:     "pointer type",
			typeStr:  "*int",
			expected: "nil",
		},
		{
			name:     "basic type",
			typeStr:  "int",
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: zeroValue requires types.Type, which is complex to construct
			// This is a placeholder for demonstrating test structure
			// Actual implementation would require proper type construction
		})
	}
}

func TestWriteAST(t *testing.T) {
	// Test for writeAST function
	// This function formats AST nodes, so we test edge cases
	t.Run("nil expression", func(t *testing.T) {
		// writeAST handles nil expressions
		// Actual test would require creating AST nodes
	})
}

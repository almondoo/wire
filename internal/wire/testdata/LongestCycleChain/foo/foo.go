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

package main

import "fmt"

func main() {
	fmt.Println(inject())
}

type A struct{}
type B struct{}
type C struct{}
type D struct{}
type E struct{}
func ProvideA(b *B) *A { return &A{} }
func ProvideB(c *C) *B { return &B{} }
func ProvideC(d *D) *C { return &C{} }
func ProvideD(e *E) *D { return &D{} }
func ProvideE(a *A) *E { return &E{} }

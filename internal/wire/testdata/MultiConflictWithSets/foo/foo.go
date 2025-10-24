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

import "github.com/almonddo/wire"

type Foo struct{}
func ProvideFoo1() *Foo { return &Foo{} }
func ProvideFoo2() *Foo { return &Foo{} }
var Set1 = wire.NewSet(ProvideFoo1)
var Set2 = wire.NewSet(ProvideFoo2)

// Copyright (c) 2022 Vincent Cheung (coolingfall@gmail.com).
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package piper

import (
	"bytes"
	"fmt"
)

type providerKey struct {
	name      string
	alias     string
	isPointer bool
}

func (k providerKey) String() string {
	buffer := new(bytes.Buffer)

	if k.isPointer {
		buffer.WriteString("*")
	}
	buffer.WriteString(k.name)
	if len(k.alias) != 0 {
		buffer.WriteString(fmt.Sprintf("(%s)", k.alias))
	}

	return buffer.String()
}

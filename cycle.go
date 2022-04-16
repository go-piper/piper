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
)

// cycleDepError represents error when cycle dependencies found in graph nodes.
type cycleDepError struct {
	nodes []*graphNode
}

func (e cycleDepError) Error() string {
	errMsgBuffer := new(bytes.Buffer)
	errMsgBuffer.WriteString("cycle dependencies found: \n\t")
	for k, v := range e.nodes {
		if k > 0 {
			errMsgBuffer.WriteString("\n\tdepends on ")
		}
		errMsgBuffer.WriteString(v.name)
	}

	return errMsgBuffer.String()
}

// cycleDependencyCheck checks if the node is cycle dependency for given dependency chain.
func cycleDependencyCheck(depChain []*graphNode, nodeToCheck *graphNode) error {
	for _, node := range depChain {
		if node != nodeToCheck {
			continue
		}

		nodes := make([]*graphNode, 0)
		nodes = append(nodes, depChain...)
		nodes = append(nodes, nodeToCheck)
		return cycleDepError{
			nodes: nodes,
		}
	}

	return nil
}

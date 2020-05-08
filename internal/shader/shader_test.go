// Copyright 2020 The Ebiten Authors
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

package shader_test

import (
	"testing"

	. "github.com/hajimehoshi/ebiten/internal/shader"
)

func TestGlsl(t *testing.T) {
	tests := []struct {
		In   string
		Dump string
	}{
		{
			In: `package main
 
type VertexOut struct {
	Position vec4 ` + "`kage:\"position\"`" + `
	TexCoord vec2
	Color    vec4
}
`,
			Dump: `var Position varying vec4 // position
var Color varying vec4
var TexCoord varying vec2
`,
		},
	}
	for _, tc := range tests {
		s, err := NewShader([]byte(tc.In))
		if err != nil {
			t.Error(err)
			continue
		}
		if got, want := s.Dump(), tc.Dump; got != want {
			t.Errorf("got: %v, want: %v", got, want)
		}
	}
}

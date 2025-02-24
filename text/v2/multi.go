// Copyright 2023 The Ebitengine Authors
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

package text

import (
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

var _ Face = (MultiFace)(nil)

// MultiFace is a Face that consists of multiple Face objects.
// The face in the first index is used in the highest priority, and the last the lowest priority.
//
// There is a known issue: if the writing directions of the faces don't agree, the rendering result might be messed up.
type MultiFace []Face

// Metrics implements Face.
func (m MultiFace) Metrics() Metrics {
	var mt Metrics
	for _, f := range m {
		mt1 := f.Metrics()
		if mt1.Height > mt.Height {
			mt.Height = mt1.Height
		}
		if mt1.HAscent > mt.HAscent {
			mt.HAscent = mt1.HAscent
		}
		if mt1.HDescent > mt.HDescent {
			mt.HDescent = mt1.HDescent
		}
		if mt1.Width > mt.Width {
			mt.Width = mt1.Width
		}
		if mt1.VAscent > mt.VAscent {
			mt.VAscent = mt1.VAscent
		}
		if mt1.VDescent > mt.VDescent {
			mt.VDescent = mt1.VDescent
		}
	}
	return mt
}

// advance implements Face.
func (m MultiFace) advance(text string) float64 {
	var a float64
	for _, c := range m.splitText(text) {
		if c.faceIndex == -1 {
			continue
		}
		f := m[c.faceIndex]
		a += f.advance(text[c.textStartIndex:c.textEndIndex])
	}
	return a
}

// hasGlyph implements Face.
func (m MultiFace) hasGlyph(r rune) bool {
	for _, f := range m {
		if f.hasGlyph(r) {
			return true
		}
	}
	return false
}

// appendGlyphsForLine implements Face.
func (m MultiFace) appendGlyphsForLine(glyphs []Glyph, line string, indexOffset int, originX, originY float64) []Glyph {
	for _, c := range m.splitText(line) {
		if c.faceIndex == -1 {
			continue
		}
		f := m[c.faceIndex]
		t := line[c.textStartIndex:c.textEndIndex]
		glyphs = f.appendGlyphsForLine(glyphs, t, indexOffset, originX, originY)
		if a := f.advance(t); f.direction().isHorizontal() {
			originX += a
		} else {
			originY += a
		}
		indexOffset += len(t)
	}
	return glyphs
}

// appendVectorPathForLine implements Face.
func (m MultiFace) appendVectorPathForLine(path *vector.Path, line string, originX, originY float64) {
	for _, c := range m.splitText(line) {
		if c.faceIndex == -1 {
			continue
		}
		f := m[c.faceIndex]
		t := line[c.textStartIndex:c.textEndIndex]
		f.appendVectorPathForLine(path, t, originX, originY)
		if a := f.advance(t); f.direction().isHorizontal() {
			originX += a
		} else {
			originY += a
		}
	}
}

// direction implements Face.
func (m MultiFace) direction() Direction {
	if len(m) == 0 {
		return DirectionLeftToRight
	}
	return m[0].direction()
}

// private implements Face.
func (m MultiFace) private() {
}

type textChunk struct {
	textStartIndex int
	textEndIndex   int
	faceIndex      int
}

func (m MultiFace) splitText(text string) []textChunk {
	var chunks []textChunk

	for ri, r := range text {
		// -1 indicates the default face index. -1 is used when no face is found for the glyph.
		fi := -1

		_, l := utf8.DecodeRuneInString(text[ri:])
		for i, f := range m {
			if !f.hasGlyph(r) {
				continue
			}
			fi = i
			break
		}

		var s int
		if len(chunks) > 0 {
			if chunks[len(chunks)-1].faceIndex == fi {
				chunks[len(chunks)-1].textEndIndex += l
				continue
			}
			s = chunks[len(chunks)-1].textEndIndex
		}
		chunks = append(chunks, textChunk{
			textStartIndex: s,
			textEndIndex:   s + l,
			faceIndex:      fi,
		})
	}

	return chunks
}

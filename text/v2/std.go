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
	"image"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var _ Face = (*StdFace)(nil)

type stdFaceGlyphImageCacheKey struct {
	rune    rune
	xoffset fixed.Int26_6

	// yoffset is always the same if the rune is the same, so this doesn't have to be a key.
}

// StdFace is a Face implementation for a semi-standard font.Face (golang.org/x/image/font).
// StdFace is useful to transit from existing codebase with text v1, or to use some bitmap fonts defined as font.Face.
// StdFace must not be copied by value.
type StdFace struct {
	f *faceWithCache

	glyphImageCache glyphImageCache[stdFaceGlyphImageCacheKey]

	addr *StdFace
}

// NewStdFace creates a new StdFace from a semi-standard font.Face.
func NewStdFace(face font.Face) *StdFace {
	s := &StdFace{
		f: &faceWithCache{
			f: face,
		},
	}
	s.addr = s
	return s
}

func (s *StdFace) copyCheck() {
	if s.addr != s {
		panic("text: illegal use of non-zero StdFace copied by value")
	}
}

// Metrics implelements Face.
func (s *StdFace) Metrics() Metrics {
	s.copyCheck()

	m := s.f.Metrics()
	return Metrics{
		Height:   fixed26_6ToFloat64(m.Height),
		HAscent:  fixed26_6ToFloat64(m.Ascent),
		HDescent: fixed26_6ToFloat64(m.Descent),
	}
}

// UnsafeInternal returns its internal font.Face.
//
// This is unsafe since this might make internal cache states out of sync.
func (s *StdFace) UnsafeInternal() font.Face {
	s.copyCheck()
	return s.f.f
}

// advance implements Face.
func (s *StdFace) advance(text string) float64 {
	return fixed26_6ToFloat64(font.MeasureString(s.f, text))
}

// hasGlyph implements Face.
func (s *StdFace) hasGlyph(r rune) bool {
	_, ok := s.f.GlyphAdvance(r)
	return ok
}

// appendGlyphsForLine implements Face.
func (s *StdFace) appendGlyphsForLine(glyphs []Glyph, line string, indexOffset int, originX, originY float64) []Glyph {
	s.copyCheck()

	origin := fixed.Point26_6{
		X: float64ToFixed26_6(originX),
		Y: float64ToFixed26_6(originY),
	}
	prevR := rune(-1)

	for i, r := range line {
		if prevR >= 0 {
			origin.X += s.f.Kern(prevR, r)
		}
		img, imgX, imgY, a := s.glyphImage(r, origin)
		if img != nil {
			// Adjust the position to the integers.
			// The current glyph images assume that they are rendered on integer positions so far.
			_, size := utf8.DecodeRuneInString(line[i:])
			glyphs = append(glyphs, Glyph{
				StartIndexInBytes: indexOffset + i,
				EndIndexInBytes:   indexOffset + i + size,
				Image:             img,
				X:                 float64(imgX),
				Y:                 float64(imgY),
			})
		}
		origin.X += a
		prevR = r
	}

	return glyphs
}

func (s *StdFace) glyphImage(r rune, origin fixed.Point26_6) (*ebiten.Image, int, int, fixed.Int26_6) {
	// Assume that StdFace's direction is always horizontal.
	origin.X = adjustGranularity(origin.X, s)
	origin.Y &^= ((1 << 6) - 1)

	b, a, _ := s.f.GlyphBounds(r)
	subpixelOffset := fixed.Point26_6{
		X: (origin.X + b.Min.X) & ((1 << 6) - 1),
		Y: (origin.Y + b.Min.Y) & ((1 << 6) - 1),
	}
	key := stdFaceGlyphImageCacheKey{
		rune:    r,
		xoffset: subpixelOffset.X,
	}
	img := s.glyphImageCache.getOrCreate(s, key, func() *ebiten.Image {
		return s.glyphImageImpl(r, subpixelOffset, b)
	})
	imgX := (origin.X + b.Min.X).Floor()
	imgY := (origin.Y + b.Min.Y).Floor()
	return img, imgX, imgY, a
}

func (s *StdFace) glyphImageImpl(r rune, subpixelOffset fixed.Point26_6, glyphBounds fixed.Rectangle26_6) *ebiten.Image {
	w, h := (glyphBounds.Max.X - glyphBounds.Min.X).Ceil(), (glyphBounds.Max.Y - glyphBounds.Min.Y).Ceil()
	if w == 0 || h == 0 {
		return nil
	}

	// Add always 1 to the size.
	// In theory, it is possible to determine whether +1 is necessary or not, but the calculation is pretty complicated.
	w++
	h++

	rgba := image.NewRGBA(image.Rect(0, 0, w, h))

	d := font.Drawer{
		Dst:  rgba,
		Src:  image.White,
		Face: s.f,
		Dot: fixed.Point26_6{
			X: -glyphBounds.Min.X + subpixelOffset.X,
			Y: -glyphBounds.Min.Y + subpixelOffset.Y,
		},
	}
	d.DrawString(string(r))

	return ebiten.NewImageFromImage(rgba)
}

// direction implelements Face.
func (s *StdFace) direction() Direction {
	return DirectionLeftToRight
}

// appendVectorPathForLine implements Face.
func (s *StdFace) appendVectorPathForLine(path *vector.Path, line string, originX, originY float64) {
}

// Metrics implelements Face.
func (s *StdFace) private() {
}

package multiface

import (
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type entry struct {
	font *sfnt.Font
	face font.Face
}

type Face struct {
	entries []entry
}

func (f *Face) AddTruetypeFace(face font.Face, fnt *sfnt.Font) {
	f.entries = append(f.entries, entry{fnt, face})
}

// AddFace adds a font.Face to the multiface.Face
func (f *Face) AddFace(face font.Face) {
	f.entries = append(f.entries, entry{nil, face})
}

func (f *Face) Close() error {
	var e error
	for i := range f.entries {
		err := f.entries[i].face.Close()
		if err != nil {
			e = err
		}
	}
	return e
}

func (f *Face) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	var b sfnt.Buffer
	for i := range f.entries {
		// sfnt.faces don't return ok = false for missing glyphs, but it seems when .Index(r) == 0 the glyph is missing
		if f.entries[i].font != nil {
			gi, err := f.entries[i].font.GlyphIndex(&b, r)
			if (gi == 0 && i < len(f.entries)-1) || err != nil {
				continue
			}
		}
		dr, mask, maskp, advance, ok = f.entries[i].face.Glyph(dot, r)
		if !ok && i < len(f.entries)-1 {
			continue
		}
		return
	}
	return
}

func (f *Face) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	var b sfnt.Buffer
	for i := range f.entries {
		// sfnt.faces don't return ok = false for missing glyphs, but it seems when .Index(r) == 0 the glyph is missing
		if f.entries[i].font != nil {
			gi, err := f.entries[i].font.GlyphIndex(&b, r)
			if (gi == 0 && i < len(f.entries)-1) || err != nil {
				continue
			}
		}
		bounds, advance, ok = f.entries[i].face.GlyphBounds(r)
		if !ok && i < len(f.entries)-1 {
			continue
		}
		return
	}
	return
}

func (f *Face) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	var b sfnt.Buffer
	for i := range f.entries {
		// sfnt.faces don't return ok = false for missing glyphs, but it seems when .Index(r) == 0 the glyph is missing
		if f.entries[i].font != nil {
			gi, err := f.entries[i].font.GlyphIndex(&b, r)
			if (gi == 0 && i < len(f.entries)-1) || err != nil {
				continue
			}
		}
		advance, ok = f.entries[i].face.GlyphAdvance(r)
		if !ok && i < len(f.entries)-1 {
			continue
		}
		return
	}
	return
}

func (f *Face) Kern(r0, r1 rune) fixed.Int26_6 {
	var b sfnt.Buffer
	for i := range f.entries {
		// sfnt.faces don't return ok = false for missing glyphs, but it seems when .Index(r) == 0 the glyph is missing
		if f.entries[i].font != nil {
			gi, err := f.entries[i].font.GlyphIndex(&b, r0)
			if (gi == 0 && i < len(f.entries)-1) || err != nil {
				continue
			}
			gi, err = f.entries[i].font.GlyphIndex(&b, r1)
			if (gi == 0 && i < len(f.entries)-1) || err != nil {
				continue

			}
		}

		var ok bool
		_, ok = f.entries[i].face.GlyphAdvance(r0)
		if !ok && i < len(f.entries)-1 {
			continue
		}
		_, ok = f.entries[i].face.GlyphAdvance(r1)
		if !ok && i < len(f.entries)-1 {
			continue
		}

		return f.entries[i].face.Kern(r0, r1)
	}
	return 0
}

func (f *Face) Metrics() font.Metrics {
	var m font.Metrics
	if len(f.entries) > 0 {
		m = f.entries[0].face.Metrics()
	}
	return m
}

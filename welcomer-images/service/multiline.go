package service

import (
	"image"
	"image/color"
	"strings"

	core "github.com/WelcomerTeam/Welcomer/welcomer-core"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	fontResize      = 0.9
	defaultFontSize = 96
)

// MultilineArguments is a list of arguments for the DrawMultiline function.
type MultilineArguments struct {
	DefaultFontSize float64 // default font size to start with

	X int
	Y int

	Width  int
	Height int

	HorizontalAlignment core.ImageAlignment
	VerticalAlignment   core.ImageAlignment

	StrokeWeight int
	StrokeColor  color.Color
	TextColor    color.Color

	Text string
}

// drawMultiline draws text using multiple lines and adds stroke.
func drawMultiline(d font.Drawer, newFace func(float64) font.Face, args MultilineArguments) error {
	lines := strings.Split(args.Text, "\n")
	if len(lines) == 0 {
		return nil
	}

	args.X += args.StrokeWeight
	args.Y += args.StrokeWeight

	args.Width -= args.StrokeWeight * 2
	args.Height -= args.StrokeWeight * 2

	var ls string

	var ll int

	d.Face = newFace(args.DefaultFontSize)

	// Calculate the widest line so we can use it as a baseline for other
	// font size calculations.
	for _, l := range lines {
		_, adv := d.BoundString(l)
		if adv.Ceil() > ll {
			ls = l
			ll = adv.Ceil()
		}
	}

	s := args.DefaultFontSize

	// If the widest line does not fit or the total line height is larger
	// than the height, we need to decrease the font size.
	if (ll > args.Width) || ((d.Face.Metrics().Height.Ceil() * len(lines)) > args.Height) {
		// Keep decreasing the font size until we can fit the width and height
		for {
			face := newFace(s)

			_, adv := font.BoundString(face, ls)

			// We will keep decreasing size until it fits, unless the font is already 1.
			// It may not fit but its likely the text will never fit after that point.
			if ((adv.Ceil() <= args.Width) && ((face.Metrics().Height.Ceil() * len(lines)) <= args.Height)) || s <= 1 {
				d.Face = face

				break
			}

			// TODO: Scale aware resizing so we do less operations
			s = float64(s) * fontResize
		}
	}

	fa := d.Face.Metrics().Ascent.Ceil()
	fh := d.Face.Metrics().Height.Ceil()
	th := len(lines) * fh

	sc := image.NewUniform(args.StrokeColor)
	tc := image.NewUniform(args.TextColor)

	for lineNo, l := range lines {
		_, adv := d.BoundString(l)

		var Dx int

		var Dy int

		switch args.HorizontalAlignment {
		case core.ImageAlignmentLeft:
			Dx = 0
		case core.ImageAlignmentCenter:
			Dx = int((args.Width - adv.Ceil()) / 2)
		case core.ImageAlignmentRight:
			Dx = args.Width - adv.Ceil()
		default:
			return ErrInvalidHorizontalAlignment
		}

		switch args.VerticalAlignment {
		case core.ImageAlignmentTopCenter:
			Dy = lineNo * fh
		case core.ImageAlignmentCenter:
			Dy = (lineNo * fh) + (args.Height / 2) - (th / 2)
		case core.ImageAlignmentBottomCenter:
			Dy = args.Height - th + (lineNo * fh)
		default:
			return ErrInvalidVerticalAlignment
		}

		Dx = args.X + Dx
		Dy = args.Y + Dy

		if args.StrokeWeight > 0 {
			p := args.StrokeWeight * args.StrokeWeight

			for dy := -args.StrokeWeight; dy <= args.StrokeWeight; dy++ {
				for dx := -args.StrokeWeight; dx <= args.StrokeWeight; dx++ {
					// Round out stroke
					if dx*dx+dy*dy >= p {
						continue
					}

					d.Dot = fixed.P(Dx+dx, Dy+dy+fa)
					d.Src = sc
					d.DrawString(l)
				}
			}
		}

		d.Dot = fixed.P(Dx, Dy+fa)
		d.Src = tc

		d.DrawString(l)
	}

	return nil
}

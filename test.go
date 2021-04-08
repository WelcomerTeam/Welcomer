package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/WelcomerTeam/WelcomerImages/pkg/multiface"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var text = "🎀 Pink-_-Panda-008 🎀\ngenji🐲\n🕙\nᶜʰⁱˡˡPizza steve®\n🤔Rock\n𝓞𝓷𝓼𝓻𝓪#0133\nA ‘ s\nдруг#1513\n👑┇𝐑𝐢𝐧𝐢 💔 𝐅𝐨𝐱™#2117\n𝑇𝑀𝑮\n!  ꧁卄卂ㄩ爪꧂ !\n-Māthîās-#1297\n𝓚𝓪𝔃𝓮꧁✞ Isa ✞꧂#5897\n[AFK] ᴇᴄʜᴏ#9098"

var fontCache = map[string]*sfnt.Font{}

func addFaceFont(face *multiface.Face, filename string, size float64, dpi float64) {
	fnt := ReadFont(filename)
	fc, _ := opentype.NewFace(fnt, &opentype.FaceOptions{Size: size, DPI: dpi})
	// if strings.HasSuffix(filename, ".ttf") {
	// 	face.AddTruetypeFace(fc, fnt)
	// } else {
	// 	face.AddFace(fc)
	// }
	face.AddTruetypeFace(fc, fnt)
}

func ReadFont(filename string) (fnt *sfnt.Font) {
	if fnt, ok := fontCache[filename]; !ok {
		data, err := ioutil.ReadFile(filename)
		checkErr(err)
		fnt, err := opentype.Parse(data)
		checkErr(err)

		fontCache[filename] = fnt

		return fnt
	} else {
		return fnt
	}
}

type Xalignment uint8
type Yalignment uint8

const (
	Left Xalignment = iota
	Middle
	Right
)

const (
	Top Yalignment = iota
	Center
	Bottom
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func drawMultiline(
	initialSize float64, d font.Drawer, new func(float64) font.Face, bounds image.Rectangle,
	x int, y int, width int, height int,
	horizontalAlignment Xalignment, verticalAlignment Yalignment, text string,
	strokeColour *image.Uniform, strokeWeight int, textColour *image.Uniform) error {

	x += strokeWeight
	y += strokeWeight
	width -= strokeWeight * 2
	height -= strokeWeight * 2

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}

	var largestString string
	var largestLength int

	d.Face = new(initialSize)

	for _, l := range lines {
		_, adv := d.BoundString(l)
		if adv.Ceil() > largestLength {
			largestString = l
			largestLength = adv.Ceil()
		}
	}

	size := initialSize
	if largestLength > bounds.Dx() {
		// Keep decreasing font size until we can fit image into maxWidth and maxHeight
		for {
			face := new(size)
			fh := face.Metrics().Height.Ceil()

			_, adv := font.BoundString(face, largestString)
			if (adv.Ceil() <= width && len(lines)*fh <= height) || size <= 1 {
				d.Face = face
				break
			}

			size = float64(size) * 0.95
		}
	}

	fa := d.Face.Metrics().Ascent.Ceil()
	fh := d.Face.Metrics().Height.Ceil()
	// Calculate total height so we can figure out vertical height.
	th := len(lines) * fh

	for lineNo, l := range lines {
		_, adv := d.BoundString(l)

		var Dx int
		var Dy int

		switch horizontalAlignment {
		case Left:
			Dx = 0
		case Middle:
			Dx = int((width - adv.Ceil()) / 2)
		case Right:
			Dx = width - adv.Ceil()
		}

		switch verticalAlignment {
		case Top:
			Dy = lineNo * fh
		case Center:
			Dy = (lineNo * fh) + (height / 2) - (th / 2)
		case Bottom:
			Dy = height - th + (lineNo * fh)
		}

		Dx = x + Dx
		Dy = y + Dy

		if strokeWeight > 0 {
			for dy := -strokeWeight; dy <= strokeWeight; dy++ {
				for dx := -strokeWeight; dx <= strokeWeight; dx++ {
					if dx*dx+dy*dy >= strokeWeight*strokeWeight {
						// give it rounded corners
						continue
					}
					d.Dot = fixed.P(Dx+dx, Dy+dy+fa)
					d.Src = strokeColour
					d.DrawString(l)
				}
			}
		}

		d.Dot = fixed.P(Dx, Dy+fa)
		d.Src = textColour
		d.DrawString(l)
	}

	return nil
}

func loadFont(size float64) font.Face {
	// Read the font data.

	face := new(multiface.Face)
	opts := &truetype.Options{Size: size, DPI: 72}

	// Gives Roboto Priority
	addFaceFont(face, "Roboto-Medium.ttf", opts.Size, opts.DPI)

	files, _ := ioutil.ReadDir("Fonts")
	for _, file := range files {
		addFaceFont(face, "Fonts/"+file.Name(), opts.Size, opts.DPI)
	}

	return face
}

func main() {
	fg, bg := image.Black, image.White
	rgba := image.NewRGBA(image.Rect(0, 0, 640, 480))
	draw.Draw(rgba, rgba.Bounds(), bg, image.Point{}, draw.Src)

	d := font.Drawer{}
	d.Dst = rgba
	d.Src = fg

	a := time.Now()
	drawMultiline(
		40, d, loadFont, rgba.Rect, 0, 0, rgba.Rect.Dx(), rgba.Rect.Dy(), Xalignment(Right), Yalignment(Bottom),
		text,
		image.NewUniform(color.Black), 4, image.NewUniform(color.White))
	println("DML", time.Since(a).Round(time.Microsecond).Milliseconds())

	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	b := bufio.NewWriter(outFile)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

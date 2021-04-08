package multiface_test

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/WelcomerTeam/WelcomerImages/pkg/multiface"
	"github.com/golang/freetype"
	"github.com/zachomedia/go-bdf"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

func TestMultiface(t *testing.T) {

	face := new(multiface.Face)

	opts := &opentype.FaceOptions{Size: 20, DPI: 96}

	var fnt *sfnt.Font
	var fc font.Face
	var err error

	// Add ArchitectsDaughter font, which does not include a glyph for ก, but has a handwriting-style glyph for a
	fnt = readFont(t, "testdata/ArchitectsDaughter-Regular.ttf")
	fc, err = opentype.NewFace(fnt, opts)
	checkErr(t, err)
	face.AddTruetypeFace(fc, fnt)

	// Add Kanit font, which does have a glyph for ก
	fnt = readFont(t, "testdata/Kanit-Regular.ttf")
	fc, err = opentype.NewFace(fnt, opts)
	checkErr(t, err)
	face.AddTruetypeFace(fc, fnt)

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	draw.Draw(img, img.Rect, image.White, image.Point{}, draw.Src)

	d := font.Drawer{}
	d.Dst = img
	d.Src = image.Black
	d.Face = face
	d.Dot = freetype.Pt(10, 25)
	d.DrawString("กa")

	f, err := os.Create("testdata/output.png")
	checkErr(t, err)
	err = png.Encode(f, img)
	checkErr(t, err)
	err = f.Close()
	checkErr(t, err)

	f, err = os.Open("testdata/reference.png")
	checkErr(t, err)
	ref, err := png.Decode(f)
	checkErr(t, err)

	if !bytes.Equal(ref.(*image.RGBA).Pix, img.Pix) {
		t.Fatal("output does not match reference")
	}
}

func TestBdf(t *testing.T) {

	face := new(multiface.Face)

	opts := &opentype.FaceOptions{Size: 20, DPI: 96}

	var fnt *sfnt.Font
	var bdffnt *bdf.Font
	var fc font.Face
	var err error

	// Add Terminus font
	bdffnt = readBdfFont(t, "testdata/ter-u12n.bdf")
	fc = &bdf.Face{Font: bdffnt}
	face.AddTruetypeFace(fc, fnt)

	// Add Kanit font, which does have a glyph for ก
	fnt = readFont(t, "testdata/Kanit-Regular.ttf")
	fc, err = opentype.NewFace(fnt, opts)
	checkErr(t, err)
	face.AddTruetypeFace(fc, fnt)

	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	draw.Draw(img, img.Rect, image.White, image.Point{}, draw.Src)

	d := font.Drawer{}
	d.Dst = img
	d.Src = image.Black
	d.Face = face
	d.Dot = freetype.Pt(10, 25)
	d.DrawString("กa")

	f, err := os.Create("testdata/output.png")
	checkErr(t, err)
	err = png.Encode(f, img)
	checkErr(t, err)
	err = f.Close()
	checkErr(t, err)

	f, err = os.Open("testdata/reference_bdf.png")
	checkErr(t, err)
	ref, err := png.Decode(f)
	checkErr(t, err)

	if !bytes.Equal(ref.(*image.RGBA).Pix, img.Pix) {
		t.Fatal("output does not match reference")
	}
}

func readFont(t *testing.T, filename string) *sfnt.Font {
	data, err := ioutil.ReadFile(filename)
	checkErr(t, err)
	fnt, err := sfnt.Parse(data)
	checkErr(t, err)
	return fnt
}

func readBdfFont(t *testing.T, filename string) *bdf.Font {
	data, err := ioutil.ReadFile(filename)
	checkErr(t, err)
	fnt, err := bdf.Parse(data)
	checkErr(t, err)
	return fnt
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

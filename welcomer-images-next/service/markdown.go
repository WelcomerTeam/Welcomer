package service

import (
	"bytes"
	"fmt"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//
// ------------------------------------------------------------
// Preprocess: support >>> block quotes
// ------------------------------------------------------------
//

var tripleQuoteRe = regexp.MustCompile(`(?m)^>>> ?`)

func preprocess(input []byte) []byte {
	return tripleQuoteRe.ReplaceAll(input, []byte("> "))
}

//
// ------------------------------------------------------------
// Discord Emoji: <a:name:id> / <:name:id>
// ------------------------------------------------------------
//

var discordEmojiRe = regexp.MustCompile(`^<(a?):(\w+):(\d+)>`)

type DiscordEmoji struct {
	ast.BaseInline
	Animated bool
	Name     string
	ID       string
}

func (n *DiscordEmoji) Kind() ast.NodeKind {
	return KindDiscordEmoji
}

func (n *DiscordEmoji) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type DiscordEmojiParser struct{}

func (p *DiscordEmojiParser) Trigger() []byte {
	return []byte{'<'}
}

func (p *DiscordEmojiParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	m := discordEmojiRe.FindSubmatch(line)
	if m == nil {
		return nil
	}

	block.Advance(len(m[0]))

	return &DiscordEmoji{
		Animated: string(m[1]) == "a",
		Name:     string(m[2]),
		ID:       string(m[3]),
	}
}

type DiscordEmojiRenderer struct{}

var KindDiscordEmoji = ast.NewNodeKind("DiscordEmoji")

func (r *DiscordEmojiRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindDiscordEmoji, r.render)
}

func (r *DiscordEmojiRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*DiscordEmoji)

	ext := "png"

	class := "d-emoji"
	if n.Animated {
		ext = "gif"
		class += " d-emoji-animated"
	}

	fmt.Fprintf(w, `<img class="%s" src="https://cdn.discordapp.com/emojis/%s.%s">`, class, n.ID, ext)

	return ast.WalkSkipChildren, nil
}

//
// ------------------------------------------------------------
// Emphasis renderer
// - *text*  → <em>
// - **text** → <u>   (SimpleMarkdown-style underline)
// ------------------------------------------------------------
//

type EmphasisRenderer struct{}

func (r *EmphasisRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindEmphasis, r.render)
}

func (r *EmphasisRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	switch n.Level {
	case 1:
		if entering {
			w.WriteString("<em>")
		} else {
			w.WriteString("</em>")
		}
	case 2:
		if entering {
			w.WriteString("<u>")
		} else {
			w.WriteString("</u>")
		}
	}

	return ast.WalkContinue, nil
}

//
// ------------------------------------------------------------
// Markdown engine
// ------------------------------------------------------------
//

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // block quotes, code blocks
		),
		goldmark.WithParserOptions(
			parser.WithInlineParsers(
				util.Prioritized(&DiscordEmojiParser{}, 600),
			),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			renderer.WithNodeRenderers(
				util.Prioritized(&DiscordEmojiRenderer{}, 600),
				util.Prioritized(&EmphasisRenderer{}, 500),
			),
		),
	)
}

//
// ------------------------------------------------------------
// Public API
// ------------------------------------------------------------
//

func Render(input string) (string, error) {
	md := newMarkdown()

	src := preprocess([]byte(input))

	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return "", err
	}

	mdString := buf.String()

	re := regexp.MustCompile(`(?s)^<p>(.*)</p>\s*$`)
	if m := re.FindStringSubmatch(mdString); m != nil {
		return m[1], nil
	}

	return mdString, nil
}

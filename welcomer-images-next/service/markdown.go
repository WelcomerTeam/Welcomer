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

	class := "emoji"
	if n.Animated {
		ext = "gif"
		class += " emoji-animated"
	}

	fmt.Fprintf(w, `<img class="%s" style="object-fit: contain; width: 1.375em; height: 1.375em; vertical-align: bottom; display: inline;" src="https://cdn.discordapp.com/emojis/%s.%s">`, class, n.ID, ext)

	return ast.WalkSkipChildren, nil
}

//
// ------------------------------------------------------------
// Emphasis renderer
// - *text*   → <em> (italic)
// - **text** → <strong> (bold)
// - __text__ → <u> (underline)
// ------------------------------------------------------------
//

type EmphasisRenderer struct{}

func (r *EmphasisRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindEmphasis, r.render)
}

func (r *EmphasisRenderer) render(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)

	// Detect delimiter type by checking the source bytes
	// We need to look backwards from the first child node to find the delimiter
	var isUnderscore bool

	// Try to find the delimiter by examining the source around the node
	if node.HasChildren() {
		firstChild := node.FirstChild()
		// Check if the first child is a Text node and has segments
		if textNode, ok := firstChild.(*ast.Text); ok && textNode.Segment.Start > 0 {
			// Look backwards in source to find the delimiter
			pos := textNode.Segment.Start - 1
			for pos >= 0 && pos > textNode.Segment.Start-3 {
				if source[pos] == '_' {
					isUnderscore = true
					break
				} else if source[pos] == '*' {
					isUnderscore = false
					break
				}
				pos--
			}
		}
	}

	switch n.Level {
	case 1:
		// Single delimiter: *text* or _text_ → italic
		if entering {
			w.WriteString("<em>")
		} else {
			w.WriteString("</em>")
		}
	case 2:
		// Double delimiter
		if isUnderscore {
			// __text__ → underline
			if entering {
				w.WriteString("<u>")
			} else {
				w.WriteString("</u>")
			}
		} else {
			// **text** → bold
			if entering {
				w.WriteString("<strong>")
			} else {
				w.WriteString("</strong>")
			}
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
			// html.WithUnsafe(),
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

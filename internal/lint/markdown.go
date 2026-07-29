package lint

import (
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

// mdParser parses Markdown to an AST. GFM adds tables, strikethrough, task
// lists, and autolinks, so table delimiters and bare URLs are classified as
// structure, not prose.
var mdParser = goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser()

// maskMarkdownProse returns one proseLine per source line, with every non-prose
// span blanked to spaces. It parses the text to a Markdown AST and un-blanks only
// the visible text nodes, so code blocks, inline code, raw HTML and <style>/CSS,
// link and image URLs, table delimiters, and YAML frontmatter never reach the
// rules. Positions are preserved: a blanked rune keeps its column, so findings
// stay path:line:col accurate. This is the AST replacement for the old
// line-based masker (see docs/decisions/0006).
func maskMarkdownProse(text string, lines []string) []proseLine {
	src := []byte(text)
	fmEnd := frontmatterEnd(src)

	orig := make([][]rune, len(lines))
	plines := make([]proseLine, len(lines))
	for i, l := range lines {
		r := []rune(l)
		orig[i] = r
		blank := make([]rune, len(r))
		for j := range blank {
			blank[j] = ' '
		}
		plines[i] = proseLine{runes: blank, line: i + 1, kind: BlockParagraph, blank: true}
	}

	lineStart := lineStartOffsets(text)

	doc := mdParser.Parse(textReader(src))
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		t, ok := n.(*ast.Text)
		if !ok {
			return ast.WalkContinue, nil
		}
		// Code-span children are code, not prose.
		if p := n.Parent(); p != nil && p.Kind() == ast.KindCodeSpan {
			return ast.WalkContinue, nil
		}
		seg := t.Segment
		if seg.Start < fmEnd { // inside frontmatter
			return ast.WalkContinue, nil
		}
		unblankSegment(seg.Start, seg.Stop, blockKindOf(n), src, lineStart, orig, plines)
		return ast.WalkContinue, nil
	})
	return plines
}

// textReader wraps source bytes for the parser.
func textReader(src []byte) text.Reader { return text.NewReader(src) }

// unblankSegment copies the original runes of a byte span back into its prose
// line and marks that line as prose of the given block kind. A Markdown text
// node never spans a line break, so the span maps to a single line.
func unblankSegment(start, stop int, kind BlockKind, src []byte, lineStart []int, orig [][]rune, plines []proseLine) {
	li := lineOf(start, lineStart)
	if li < 0 || li >= len(plines) {
		return
	}
	runeStart := utf8.RuneCount(src[lineStart[li]:start])
	runeLen := utf8.RuneCount(src[start:stop])
	line := orig[li]
	if runeStart >= len(line) {
		return
	}
	end := runeStart + runeLen
	if end > len(line) {
		end = len(line)
	}
	copy(plines[li].runes[runeStart:end], line[runeStart:end])
	plines[li].blank = false
	if kind != BlockParagraph {
		plines[li].kind = kind
	}
}

// blockKindOf maps a node to a sentence block kind by its nearest block ancestor:
// heading text is a heading, list-item text is a list item, everything else is a
// paragraph.
func blockKindOf(n ast.Node) BlockKind {
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p.Kind() {
		case ast.KindHeading:
			return BlockHeading
		case ast.KindListItem:
			return BlockListItem
		}
	}
	return BlockParagraph
}

// lineStartOffsets returns the byte offset of each line start in text (split on
// "\n"), so a byte position maps back to a line.
func lineStartOffsets(text string) []int {
	out := []int{0}
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			out = append(out, i+1)
		}
	}
	return out
}

// lineOf returns the index of the line that contains byte offset pos.
func lineOf(pos int, lineStart []int) int {
	lo, hi := 0, len(lineStart)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if lineStart[mid] <= pos {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// frontmatterEnd returns the byte offset just past a leading YAML frontmatter
// block (--- ... --- or ... terminator), or 0 when there is none.
func frontmatterEnd(src []byte) int {
	s := string(src)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return 0
	}
	nl := strings.IndexByte(s, '\n')
	off := nl + 1
	for off < len(s) {
		end := strings.IndexByte(s[off:], '\n')
		var line string
		if end < 0 {
			line = s[off:]
		} else {
			line = s[off : off+end]
		}
		if t := strings.TrimRight(line, "\r"); t == "---" || t == "..." {
			if end < 0 {
				return len(s)
			}
			return off + end + 1
		}
		if end < 0 {
			break
		}
		off += end + 1
	}
	return 0
}

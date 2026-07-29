package lint

import (
	"strings"
	"testing"
)

func TestParseSimpleSentence(t *testing.T) {
	doc := Parse("t.txt", "Hello world.", ParseOptions{})
	if doc.Path != "t.txt" {
		t.Errorf("Path = %q", doc.Path)
	}
	if doc.Text != "Hello world." {
		t.Errorf("Text = %q", doc.Text)
	}
	if len(doc.Lines) != 1 || doc.Lines[0] != "Hello world." {
		t.Fatalf("Lines = %#v", doc.Lines)
	}
	if len(doc.Sentences) != 1 {
		t.Fatalf("Sentences = %d, want 1", len(doc.Sentences))
	}
	s := doc.Sentences[0]
	if s.Block != BlockParagraph {
		t.Errorf("Block = %d, want paragraph", s.Block)
	}
	if s.WordCount() != 2 {
		t.Errorf("WordCount = %d, want 2", s.WordCount())
	}
	if len(s.Tokens) != 3 {
		t.Fatalf("Tokens = %d, want 3", len(s.Tokens))
	}
	want := []Token{
		{Text: "Hello", Lower: "hello", Kind: KindWord, Line: 1, Col: 1, EndLine: 1, EndCol: 6},
		{Text: "world", Lower: "world", Kind: KindWord, Line: 1, Col: 7, EndLine: 1, EndCol: 12},
		{Text: ".", Lower: ".", Kind: KindPunct, Line: 1, Col: 12, EndLine: 1, EndCol: 13},
	}
	for i, w := range want {
		if s.Tokens[i] != w {
			t.Errorf("token %d = %#v, want %#v", i, s.Tokens[i], w)
		}
	}
}

func TestParseTokenKinds(t *testing.T) {
	doc := Parse("t.txt", "Set 3 bolts.", ParseOptions{})
	toks := doc.Sentences[0].Tokens
	if len(toks) != 4 {
		t.Fatalf("tokens = %d, want 4", len(toks))
	}
	if toks[0].Kind != KindWord || !toks[0].IsWord() {
		t.Errorf("token 0 kind = %d", toks[0].Kind)
	}
	if toks[1].Kind != KindNumber || toks[1].Text != "3" {
		t.Errorf("token 1 = %#v, want number 3", toks[1])
	}
	if toks[1].IsWord() {
		t.Errorf("number should not be IsWord")
	}
	if toks[3].Kind != KindPunct || toks[3].Text != "." {
		t.Errorf("token 3 = %#v, want punct", toks[3])
	}
}

func TestParseHyphenAndApostropheJoin(t *testing.T) {
	doc := Parse("t.txt", "It's a well-made part.", ParseOptions{})
	toks := doc.Sentences[0].Tokens
	// It's a well-made part .
	wantTexts := []string{"It's", "a", "well-made", "part", "."}
	if len(toks) != len(wantTexts) {
		t.Fatalf("tokens = %d (%v), want %d", len(toks), tokenTexts(toks), len(wantTexts))
	}
	for i, w := range wantTexts {
		if toks[i].Text != w {
			t.Errorf("token %d = %q, want %q", i, toks[i].Text, w)
		}
	}
}

func TestParseTrailingConnectorNotJoined(t *testing.T) {
	// A hyphen not followed by a word rune is its own punctuation token.
	doc := Parse("t.txt", "end- stop", ParseOptions{})
	toks := doc.Sentences[0].Tokens
	wantTexts := []string{"end", "-", "stop"}
	if len(toks) != len(wantTexts) {
		t.Fatalf("tokens = %v, want %v", tokenTexts(toks), wantTexts)
	}
	for i, w := range wantTexts {
		if toks[i].Text != w {
			t.Errorf("token %d = %q, want %q", i, toks[i].Text, w)
		}
	}
}

func TestParseMultibyteColumns(t *testing.T) {
	// The é is a single rune; columns count runes, not bytes.
	doc := Parse("t.txt", "café ready", ParseOptions{})
	toks := doc.Sentences[0].Tokens
	if len(toks) != 2 {
		t.Fatalf("tokens = %v", tokenTexts(toks))
	}
	if toks[0].Text != "café" || toks[0].Col != 1 || toks[0].EndCol != 5 {
		t.Errorf("token 0 = %#v, want café at 1..5", toks[0])
	}
	if toks[1].Col != 6 {
		t.Errorf("token 1 Col = %d, want 6", toks[1].Col)
	}
}

func TestParseSentenceSplitting(t *testing.T) {
	doc := Parse("t.txt", "First one. Second one.", ParseOptions{})
	if len(doc.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(doc.Sentences))
	}
	if doc.Sentences[0].WordCount() != 2 || doc.Sentences[1].WordCount() != 2 {
		t.Errorf("word counts = %d, %d", doc.Sentences[0].WordCount(), doc.Sentences[1].WordCount())
	}
	if doc.Sentences[1].Line != 1 || doc.Sentences[1].Col != 12 {
		t.Errorf("second sentence at %d:%d, want 1:12", doc.Sentences[1].Line, doc.Sentences[1].Col)
	}
}

func TestParseAbbreviationDoesNotSplit(t *testing.T) {
	// "e.g." is lower case after the dots, so it does not end the sentence.
	doc := Parse("t.txt", "Use it, e.g. now here.", ParseOptions{})
	if len(doc.Sentences) != 1 {
		t.Fatalf("sentences = %d, want 1", len(doc.Sentences))
	}
}

func TestParseCapitalAfterPeriodSplits(t *testing.T) {
	doc := Parse("t.txt", "Do this. Then that.", ParseOptions{})
	if len(doc.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(doc.Sentences))
	}
}

func TestParseNumberAfterPeriodSplits(t *testing.T) {
	doc := Parse("t.txt", "See this. 3 more.", ParseOptions{})
	if len(doc.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(doc.Sentences))
	}
}

func TestParseExclamationAndQuestionAlwaysSplit(t *testing.T) {
	doc := Parse("t.txt", "Stop! go? done.", ParseOptions{})
	if len(doc.Sentences) != 3 {
		t.Fatalf("sentences = %d, want 3", len(doc.Sentences))
	}
}

func TestParseMarkdownHeading(t *testing.T) {
	doc := Parse("t.md", "# Title here", ParseOptions{Markdown: true})
	if len(doc.Sentences) != 1 {
		t.Fatalf("sentences = %d, want 1", len(doc.Sentences))
	}
	s := doc.Sentences[0]
	if s.Block != BlockHeading {
		t.Errorf("Block = %d, want heading", s.Block)
	}
	if s.WordCount() != 2 {
		t.Errorf("WordCount = %d, want 2", s.WordCount())
	}
	if s.Tokens[0].Text != "Title" || s.Tokens[0].Col != 3 {
		t.Errorf("first token = %#v, want Title at col 3", s.Tokens[0])
	}
}

func TestParseMarkdownHeadingRequiresSpace(t *testing.T) {
	// "#Title" has no space, so it is not a heading; it is a paragraph.
	doc := Parse("t.md", "#Title here", ParseOptions{Markdown: true})
	if len(doc.Sentences) != 1 {
		t.Fatalf("sentences = %d", len(doc.Sentences))
	}
	if doc.Sentences[0].Block != BlockParagraph {
		t.Errorf("Block = %d, want paragraph", doc.Sentences[0].Block)
	}
	// The leading '#' is punctuation, kept in the token stream.
	if doc.Sentences[0].Tokens[0].Text != "#" {
		t.Errorf("first token = %q, want #", doc.Sentences[0].Tokens[0].Text)
	}
}

func TestParseMarkdownListItem(t *testing.T) {
	doc := Parse("t.md", "- Do the thing", ParseOptions{Markdown: true})
	if len(doc.Sentences) != 1 {
		t.Fatalf("sentences = %d, want 1", len(doc.Sentences))
	}
	s := doc.Sentences[0]
	if s.Block != BlockListItem {
		t.Errorf("Block = %d, want list item", s.Block)
	}
	if s.Tokens[0].Text != "Do" || s.Tokens[0].Col != 3 {
		t.Errorf("first token = %#v, want Do at col 3", s.Tokens[0])
	}
}

func TestParseMarkdownNumberedListItem(t *testing.T) {
	doc := Parse("t.md", "1. Open the valve", ParseOptions{Markdown: true})
	if len(doc.Sentences) != 1 || doc.Sentences[0].Block != BlockListItem {
		t.Fatalf("sentences = %#v", doc.Sentences)
	}
	if doc.Sentences[0].Tokens[0].Text != "Open" || doc.Sentences[0].Tokens[0].Col != 4 {
		t.Errorf("first token = %#v, want Open at col 4", doc.Sentences[0].Tokens[0])
	}
}

func TestParseMarkdownFencedCodeHidden(t *testing.T) {
	text := "Before text.\n\n```\ndont panic here\n```\n\nAfter text."
	doc := Parse("t.md", text, ParseOptions{Markdown: true})
	// Only the two prose paragraphs become sentences; the fence content is hidden.
	if len(doc.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2: %v", len(doc.Sentences), sentenceTexts(doc))
	}
	for _, s := range doc.Sentences {
		for _, tok := range s.Tokens {
			if tok.Lower == "panic" {
				t.Errorf("fenced code leaked token %q", tok.Text)
			}
		}
	}
}

func TestParseMarkdownInlineCodeHidden(t *testing.T) {
	doc := Parse("t.md", "Run `dont` now.", ParseOptions{Markdown: true})
	for _, s := range doc.Sentences {
		for _, tok := range s.Tokens {
			if tok.Lower == "dont" {
				t.Errorf("inline code leaked token %q", tok.Text)
			}
		}
	}
}

func TestParseMarkdownLinkDestinationHidden(t *testing.T) {
	doc := Parse("t.md", "See [the guide](http://dont.example.com) now.", ParseOptions{Markdown: true})
	var texts []string
	for _, s := range doc.Sentences {
		for _, tok := range s.Tokens {
			texts = append(texts, tok.Lower)
			if tok.Lower == "dont" || tok.Lower == "http" {
				t.Errorf("link destination leaked %q", tok.Text)
			}
		}
	}
	// The visible link text remains.
	if !sliceHas(texts, "guide") {
		t.Errorf("link text missing; tokens = %v", texts)
	}
}

func TestParseMarkdownOffKeepsCodeFence(t *testing.T) {
	// With markdown off, a fence line is just text.
	doc := Parse("t.txt", "```\nhello world\n```", ParseOptions{Markdown: false})
	found := false
	for _, s := range doc.Sentences {
		for _, tok := range s.Tokens {
			if tok.Lower == "hello" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("markdown-off dropped fenced text")
	}
}

func TestParseBlankLineSeparatesBlocks(t *testing.T) {
	doc := Parse("t.txt", "First block\n\nSecond block", ParseOptions{})
	if len(doc.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2", len(doc.Sentences))
	}
	if doc.Sentences[1].Line != 3 {
		t.Errorf("second block line = %d, want 3", doc.Sentences[1].Line)
	}
}

func TestParseCRLFStripped(t *testing.T) {
	doc := Parse("t.txt", "line one\r\nline two", ParseOptions{})
	if doc.Lines[0] != "line one" {
		t.Errorf("Lines[0] = %q, carriage return not stripped", doc.Lines[0])
	}
}

func TestParseEmptyText(t *testing.T) {
	doc := Parse("t.txt", "", ParseOptions{})
	if len(doc.Sentences) != 0 {
		t.Errorf("sentences = %d, want 0", len(doc.Sentences))
	}
	if len(doc.Lines) != 1 {
		t.Errorf("Lines = %d, want 1", len(doc.Lines))
	}
}

func TestWordCountIgnoresPunct(t *testing.T) {
	doc := Parse("t.txt", "one, two; three.", ParseOptions{})
	if got := doc.Sentences[0].WordCount(); got != 3 {
		t.Errorf("WordCount = %d, want 3", got)
	}
}

// helpers

func tokenTexts(toks []Token) []string {
	out := make([]string, len(toks))
	for i, t := range toks {
		out[i] = t.Text
	}
	return out
}

func sentenceTexts(doc *Document) []string {
	var out []string
	for _, s := range doc.Sentences {
		out = append(out, tokenTexts(s.Tokens)...)
	}
	return out
}

func sliceHas(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// FuzzParse fuzzes lint.Parse with markdown on and off. Parse must never panic
// and must keep token positions inside the source lines.
func FuzzParse(f *testing.F) {
	seeds := []string{
		"",
		"Hello world.",
		"# Heading\n\n- item one\n- item two",
		"```\ncode\n```\ntext",
		"It's a well-made café.",
		"See [x](http://y) and `z`.",
		"multi\nline\n\nparagraph text here",
		"1. step one\n2. step two; and then stop",
		"\r\n\t   \n#no space heading",
		"a.b.c.d e.g. i.e. etc.",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, text string) {
		for _, md := range []bool{false, true} {
			doc := Parse("fuzz.md", text, ParseOptions{Markdown: md})
			if doc == nil {
				t.Fatal("Parse returned nil")
			}
			for _, s := range doc.Sentences {
				_ = s.WordCount()
				for _, tok := range s.Tokens {
					if tok.Col < 1 || tok.EndCol < tok.Col {
						t.Fatalf("bad token span %#v for input %q md=%v", tok, text, md)
					}
					if tok.Line < 1 || tok.Line > len(doc.Lines) {
						t.Fatalf("token line %d out of range (lines=%d) input %q", tok.Line, len(doc.Lines), text)
					}
				}
			}
		}
	})
}

func TestParseMarkdownHTMLBlockAndCSSHidden(t *testing.T) {
	text := "Real prose here.\n\n<style>\n.card { color: #ffffff; margin: 0; }\n</style>\n\nMore prose."
	doc := Parse("t.md", text, ParseOptions{Markdown: true})
	for _, tok := range sentenceTexts(doc) {
		for _, leak := range []string{"style", "card", "color", "ffffff", "margin"} {
			if strings.EqualFold(tok, leak) {
				t.Errorf("CSS/HTML leaked token %q", tok)
			}
		}
	}
	if !sliceHas(sentenceTexts(doc), "Real") || !sliceHas(sentenceTexts(doc), "prose") {
		t.Errorf("real prose dropped: %v", sentenceTexts(doc))
	}
}

func TestParseMarkdownFrontmatterHidden(t *testing.T) {
	text := "---\ntitle: My Doc\ndescription: dont lint this\n---\n\nActual sentence."
	doc := Parse("t.md", text, ParseOptions{Markdown: true})
	for _, tok := range sentenceTexts(doc) {
		if strings.EqualFold(tok, "title") || strings.EqualFold(tok, "dont") || strings.EqualFold(tok, "description") {
			t.Errorf("frontmatter leaked token %q; tokens=%v", tok, sentenceTexts(doc))
		}
	}
	if !sliceHas(sentenceTexts(doc), "Actual") {
		t.Errorf("body prose dropped: %v", sentenceTexts(doc))
	}
}

func TestParseMarkdownInlineHTMLTagStrippedKeepsText(t *testing.T) {
	doc := Parse("t.md", "A word <b>kept</b> here.", ParseOptions{Markdown: true})
	texts := sentenceTexts(doc)
	if !sliceHas(texts, "kept") {
		t.Errorf("inline HTML inner text dropped: %v", texts)
	}
	if sliceHas(texts, "b") {
		t.Errorf("inline HTML tag leaked: %v", texts)
	}
}

func TestParseMarkdownTableCellsAreProseNotDelimiters(t *testing.T) {
	text := "| Head | Note |\n|------|------|\n| open | valve |"
	doc := Parse("t.md", text, ParseOptions{Markdown: true})
	texts := sentenceTexts(doc)
	if !sliceHas(texts, "valve") {
		t.Errorf("table cell text dropped: %v", texts)
	}
	for _, tok := range texts {
		if strings.Contains(tok, "|") || strings.Contains(tok, "---") {
			t.Errorf("table delimiter leaked as token %q", tok)
		}
	}
}

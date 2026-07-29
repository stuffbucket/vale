package lint

import (
	"strings"
	"unicode"
)

// TokenKind tells what kind of text a token holds.
type TokenKind int

const (
	// KindWord is a run of letters, with optional inner apostrophes or hyphens.
	KindWord TokenKind = iota
	// KindNumber is a run that starts with a digit.
	KindNumber
	// KindPunct is a single mark of punctuation.
	KindPunct
)

// Token is one unit of text with its position. The line and column numbers
// start at 1. The column counts runes. Col points at the first rune. EndCol
// points at the position just after the last rune.
type Token struct {
	Text    string
	Lower   string
	Kind    TokenKind
	Line    int
	Col     int
	EndLine int
	EndCol  int
}

// IsWord tells if the token is a word.
func (t Token) IsWord() bool { return t.Kind == KindWord }

// BlockKind tells what kind of block a sentence comes from.
type BlockKind int

const (
	// BlockParagraph is normal prose.
	BlockParagraph BlockKind = iota
	// BlockHeading is a Markdown heading line.
	BlockHeading
	// BlockListItem is one item in a list.
	BlockListItem
)

// Sentence is a run of tokens that forms one sentence. It keeps the block kind
// so that rules can change their behavior for headings and list items.
type Sentence struct {
	Tokens []Token
	Block  BlockKind
	Line   int
	Col    int
}

// WordCount gives the number of word and number tokens in the sentence.
func (s Sentence) WordCount() int {
	n := 0
	for _, t := range s.Tokens {
		if t.Kind == KindWord || t.Kind == KindNumber {
			n++
		}
	}
	return n
}

// Document is a parsed source text. It holds the raw text, the lines, and the
// sentences that the tokenizer found.
type Document struct {
	Path      string
	Text      string
	Lines     []string
	Sentences []Sentence
}

// ParseOptions controls how the tokenizer reads a document.
type ParseOptions struct {
	// Markdown tells the tokenizer to hide code blocks, inline code, and link
	// destinations, and to find headings and list items.
	Markdown bool
}

// Parse reads text into a Document. The path is only for reports.
func Parse(path, text string, opts ParseOptions) *Document {
	doc := &Document{Path: path, Text: text}
	doc.Lines = splitLines(text)
	var prose []proseLine
	if opts.Markdown {
		prose = maskMarkdownProse(text, doc.Lines)
	} else {
		prose = maskLines(doc.Lines)
	}
	blocks := groupBlocks(prose, opts.Markdown)
	for _, b := range blocks {
		toks := tokenizeBlock(b)
		doc.Sentences = append(doc.Sentences, splitSentences(toks, b.kind)...)
	}
	return doc
}

// splitLines splits text into lines. It drops the line-feed and any carriage
// return, but keeps empty lines so that positions stay correct.
func splitLines(text string) []string {
	raw := strings.Split(text, "\n")
	out := make([]string, len(raw))
	for i, line := range raw {
		out[i] = strings.TrimSuffix(line, "\r")
	}
	return out
}

// proseLine is one source line after the tokenizer hides code. The runes slice
// has the same length as the source line in runes; hidden runes become spaces.
type proseLine struct {
	runes []rune
	line  int // 1-based source line number
	kind  BlockKind
	blank bool // the source line held only structure or space
}

// maskLines turns source lines into prose lines for non-Markdown text: every
// line is prose as written. Markdown goes through maskMarkdownProse instead
// (AST-based), so the line heuristics that used to live here are gone.
func maskLines(lines []string) []proseLine {
	out := make([]proseLine, 0, len(lines))
	for i, src := range lines {
		runes := []rune(src)
		out = append(out, proseLine{
			runes: append([]rune(nil), runes...),
			line:  i + 1,
			kind:  BlockParagraph,
			blank: strings.TrimSpace(src) == "",
		})
	}
	return out
}

// block is a run of prose lines that forms one unit of text.
type block struct {
	lines []proseLine
	kind  BlockKind
}

// groupBlocks joins prose lines into blocks. A blank line ends a block. A
// heading or a list item is its own block.
func groupBlocks(lines []proseLine, markdown bool) []block {
	var blocks []block
	var cur *block
	flush := func() {
		if cur != nil && len(cur.lines) > 0 {
			blocks = append(blocks, *cur)
		}
		cur = nil
	}
	for _, pl := range lines {
		if pl.blank {
			flush()
			continue
		}
		if markdown && (pl.kind == BlockHeading || pl.kind == BlockListItem) {
			flush()
			blocks = append(blocks, block{lines: []proseLine{pl}, kind: pl.kind})
			continue
		}
		if cur == nil {
			cur = &block{kind: BlockParagraph}
		}
		cur.lines = append(cur.lines, pl)
	}
	flush()
	return blocks
}

// tokenizeBlock turns a block into a flat slice of tokens with positions.
func tokenizeBlock(b block) []Token {
	var toks []Token
	for _, pl := range b.lines {
		toks = append(toks, tokenizeLine(pl.runes, pl.line)...)
	}
	return toks
}

// isConnector tells if a rune can join two word characters inside one word.
func isConnector(r rune) bool {
	return r == '\'' || r == '’' || r == '-'
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// tokenizeLine splits one prose line into tokens. Column numbers start at 1 and
// count runes.
func tokenizeLine(runes []rune, line int) []Token {
	var toks []Token
	i := 0
	n := len(runes)
	for i < n {
		r := runes[i]
		if r == ' ' || r == '\t' {
			i++
			continue
		}
		if isWordRune(r) {
			start := i
			i++
			for i < n {
				if isWordRune(runes[i]) {
					i++
					continue
				}
				// A connector joins only when a word rune follows it.
				if isConnector(runes[i]) && i+1 < n && isWordRune(runes[i+1]) {
					i++
					continue
				}
				break
			}
			text := string(runes[start:i])
			kind := KindWord
			if unicode.IsDigit(runes[start]) {
				kind = KindNumber
			}
			toks = append(toks, Token{
				Text: text, Lower: strings.ToLower(text), Kind: kind,
				Line: line, Col: start + 1, EndLine: line, EndCol: i + 1,
			})
			continue
		}
		// Any other rune is a single punctuation token.
		toks = append(toks, Token{
			Text: string(r), Lower: string(r), Kind: KindPunct,
			Line: line, Col: i + 1, EndLine: line, EndCol: i + 2,
		})
		i++
	}
	return toks
}

// splitSentences groups tokens into sentences. It ends a sentence at a period,
// a question mark, or an exclamation mark, when the mark truly ends a sentence.
// Headings and list items become a single sentence each.
func splitSentences(toks []Token, kind BlockKind) []Sentence {
	if len(toks) == 0 {
		return nil
	}
	if kind == BlockHeading {
		return []Sentence{makeSentence(toks, kind)}
	}
	var sentences []Sentence
	start := 0
	for i, t := range toks {
		if t.Kind != KindPunct {
			continue
		}
		if t.Text != "." && t.Text != "!" && t.Text != "?" {
			continue
		}
		if !endsSentence(toks, i) {
			continue
		}
		sentences = append(sentences, makeSentence(toks[start:i+1], kind))
		start = i + 1
	}
	if start < len(toks) {
		sentences = append(sentences, makeSentence(toks[start:], kind))
	}
	return sentences
}

// endsSentence decides if the punctuation token at index i truly ends a
// sentence. A period ends a sentence when it is the last token, or when the
// next word starts with a capital letter. This keeps groups like "e.g." and
// version numbers together.
func endsSentence(toks []Token, i int) bool {
	if toks[i].Text != "." {
		return true
	}
	// Find the next word or number token.
	for j := i + 1; j < len(toks); j++ {
		if toks[j].Kind == KindPunct {
			if toks[j].Text == "." || toks[j].Text == "!" || toks[j].Text == "?" {
				return false
			}
			continue
		}
		first := []rune(toks[j].Text)[0]
		return unicode.IsUpper(first) || unicode.IsDigit(first)
	}
	return true
}

func makeSentence(toks []Token, kind BlockKind) Sentence {
	s := Sentence{Tokens: toks, Block: kind}
	if len(toks) > 0 {
		s.Line = toks[0].Line
		s.Col = toks[0].Col
	}
	return s
}

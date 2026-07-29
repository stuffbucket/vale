// Package slop holds the curated watchlists for the STE.Slop* rule family:
// evidence-backed markers of LLM-generated "slop." The lists are deliberately
// small, closed, and versioned (markers decay over time), and the data lives
// here so both the rules and the eval harness can share it. See
// research/ai-slop/ for the evidence and the false-positive caveats.
package slop

// Family names a model family a marker most signals. It is descriptive only —
// a match is a plain-language nudge, never an authorship claim.
const (
	FamilyOpenAI = "OpenAI"
	FamilyClaude = "Anthropic"
	FamilyShared = "cross-family"
)

// SlopWords maps a lower-case slop word to the model family it most signals.
// The set is the low-baseline "spike" vocabulary that is also non-plain diction
// STE discourages on independent grounds (research/ai-slop/root-causes/
// lexical-spikes.md and model-signatures.md). Kept closed and small.
var SlopWords = map[string]string{
	// OpenAI / ChatGPT-era excess vocabulary (Kobak et al. 2406.07016).
	"delve": FamilyOpenAI, "delves": FamilyOpenAI, "delving": FamilyOpenAI,
	"delved": FamilyOpenAI, "underscore": FamilyOpenAI, "underscores": FamilyOpenAI,
	"showcasing": FamilyOpenAI, "showcase": FamilyOpenAI, "intricate": FamilyOpenAI,
	"meticulous": FamilyOpenAI, "meticulously": FamilyOpenAI, "pivotal": FamilyOpenAI,
	"realm": FamilyOpenAI, "boasts": FamilyOpenAI, "garner": FamilyOpenAI,
	"interplay": FamilyOpenAI, "tapestry": FamilyOpenAI, "testament": FamilyOpenAI,
	"nestled": FamilyOpenAI, "elucidate": FamilyOpenAI, "groundbreaking": FamilyOpenAI,
	"seamlessly": FamilyOpenAI, "unparalleled": FamilyOpenAI, "multifaceted": FamilyOpenAI,
	// Cross-family elevated register.
	"leverage": FamilyShared, "harness": FamilyShared, "vibrant": FamilyShared,
	// Anthropic / Claude-associated diction (community-observed, lower confidence).
	"nuanced": FamilyClaude, "genuinely": FamilyClaude,
}

// EvaluativeAdjectives are positive-evaluative praise words whose elevated
// DENSITY (not any single hit) is the signal (2403.07183, 2403.16887). Used by
// the density rule, never as a per-hit flag.
var EvaluativeAdjectives = []string{
	"commendable", "meticulous", "notable", "noteworthy", "innovative",
	"versatile", "comprehensive", "invaluable", "robust", "remarkable",
	"exceptional", "thorough", "seamless", "cutting-edge",
}

// Hedges are single-word hedges. Individual hedges are legitimate; only
// clustering within one sentence is a signal (2509.24202). Used by the density
// rule.
var Hedges = []string{
	"may", "might", "could", "likely", "possibly", "perhaps", "presumably",
	"seems", "appears", "typically", "generally", "probably", "arguably",
}

// RestatementMarkers are phrases that announce a restatement — a proxy for a
// model reinventing the same point (research/ai-slop/root-causes/
// repetition-and-reinvention.md). Each is a token sequence (lower case).
var RestatementMarkers = [][]string{
	{"in", "other", "words"},
	{"to", "put", "it", "another", "way"},
	{"that", "is", "to", "say"},
	{"to", "reiterate"},
	{"as", "mentioned"},
	{"as", "noted", "above"},
	{"in", "essence"},
	{"simply", "put"},
}

// ImpersonalHedgePatterns are impersonal-it modal padding templates. Each entry
// is a per-position option list: a token matches when it is one of the options
// at that position (2509.24202).
var ImpersonalHedgePatterns = [][][]string{
	{{"it"}, {"can", "could", "may", "might"}, {"be"}, {"argued", "said", "assumed", "shown", "noted", "seen"}, {"that"}},
	{{"it"}, {"is"}, {"possible"}, {"that"}},
	{{"there"}, {"is"}, {"a"}, {"possibility"}, {"that"}},
}

// NegativeParallelismPatterns match the "not only X but also Y" / "not just X
// but Y" construction as a two-anchor token check: the first anchor, then the
// second later in the same sentence (Wikipedia: Signs of AI writing).
var NegativeParallelismPatterns = []struct{ First, Second []string }{
	{First: []string{"not", "only"}, Second: []string{"but", "also"}},
	{First: []string{"not", "just"}, Second: []string{"but"}},
}

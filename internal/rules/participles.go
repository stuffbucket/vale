package rules

import "strings"

// beVerbs is the set of forms of the verb "be". A passive verb uses one of
// these forms with a past participle.
var beVerbs = map[string]bool{
	"am": true, "is": true, "are": true, "was": true, "were": true,
	"be": true, "been": true, "being": true,
}

// irregularParticiples is a set of common past participles that do not end in
// "ed". The passive-voice rule uses it together with the "ed" test.
var irregularParticiples = map[string]bool{
	"beaten": true, "become": true, "begun": true, "bent": true, "bound": true,
	"broken": true, "brought": true, "built": true, "bought": true, "caught": true,
	"chosen": true, "come": true, "cut": true, "dealt": true, "done": true,
	"drawn": true, "driven": true, "eaten": true, "fallen": true, "fed": true,
	"felt": true, "found": true, "frozen": true, "given": true, "gone": true,
	"grown": true, "held": true, "hidden": true, "hit": true, "hurt": true,
	"kept": true, "known": true, "laid": true, "led": true, "left": true,
	"lost": true, "made": true, "meant": true, "met": true, "paid": true,
	"put": true, "read": true, "run": true, "said": true, "seen": true,
	"sent": true, "set": true, "shown": true, "shut": true, "sold": true,
	"spent": true, "split": true, "spread": true, "sung": true, "taken": true,
	"taught": true, "told": true, "thrown": true, "understood": true, "worn": true,
	"written": true,
}

// nonParticipleED holds words that end in "ed" but are not past participles.
// They stop the passive-voice rule from making false findings.
var nonParticipleED = map[string]bool{
	"need": true, "indeed": true, "speed": true, "proceed": true, "exceed": true,
	"embed": true, "feed": true, "seed": true, "bed": true, "red": true,
	"bred": true, "shed": true, "sled": true, "shred": true, "wed": true,
	"fled": true, "bled": true, "sped": true, "hundred": true, "sacred": true,
}

// looksLikeParticiple tells if a lowercase word can be a past participle.
func looksLikeParticiple(word string) bool {
	if irregularParticiples[word] {
		return true
	}
	if nonParticipleED[word] {
		return false
	}
	if strings.HasSuffix(word, "ed") && len([]rune(word)) >= 4 {
		return true
	}
	return false
}

package eval

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteText prints a compact human report: families ranked by slop density, then
// the per-model rows.
func WriteText(w io.Writer, rep *Report) {
	fmt.Fprintf(w, "endpoint: %s\n\n", rep.Endpoint)
	fmt.Fprintln(w, "family slop density (STE.Slop findings per 100 words), most slop first:")
	for _, f := range rep.Families {
		fmt.Fprintf(w, "  %-12s slop %6.2f  all %6.2f  (%d models, %d words)\n",
			f.Family, f.SlopPer100Words, f.TotalPer100Word, f.Models, f.Words)
	}
	fmt.Fprintln(w, "\nper model:")
	for _, m := range rep.Models {
		note := ""
		if m.Errors > 0 {
			note = fmt.Sprintf("  [%d errors: %s]", m.Errors, m.FirstError)
		}
		fmt.Fprintf(w, "  %-28s %-10s slop %6.2f  all %6.2f  (%d words)%s\n",
			m.Model, m.Family, m.SlopPer100Words, m.TotalPer100Word, m.Words, note)
	}
}

// WriteJSON writes the aggregated results (without the raw sample text) as JSON.
func WriteJSON(w io.Writer, rep *Report) error {
	out := struct {
		Endpoint string         `json:"endpoint"`
		Families []FamilyResult `json:"families"`
		Models   []ModelResult  `json:"models"`
	}{Endpoint: rep.Endpoint, Families: rep.Families, Models: rep.Models}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

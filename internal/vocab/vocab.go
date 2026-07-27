// Package vocab holds the Simplified Technical English vocabulary data. The
// Substitutions slice is generated from the vendored OpenSTE wordset by the
// generator package. Do not edit the generated file by hand; run "vale gen".
package vocab

//go:generate go run github.com/stuffbucket/vale/cmd/vale gen

// Substitution is one unapproved word or phrase and the approved words that can
// take its place.
type Substitution struct {
	// Word is the unapproved word or phrase, in lower case.
	Word string
	// Alternatives lists the approved replacements, in a stable order.
	Alternatives []string
}

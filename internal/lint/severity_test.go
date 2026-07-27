package lint

import "testing"

func TestSeverityValid(t *testing.T) {
	tests := []struct {
		name string
		sev  Severity
		want bool
	}{
		{"error", SeverityError, true},
		{"warning", SeverityWarning, true},
		{"suggestion", SeveritySuggestion, true},
		{"empty", Severity(""), false},
		{"unknown", Severity("critical"), false},
		{"wrongcase", Severity("Error"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sev.Valid(); got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverityAtLeast(t *testing.T) {
	tests := []struct {
		name  string
		sev   Severity
		other Severity
		want  bool
	}{
		{"error>=error", SeverityError, SeverityError, true},
		{"error>=warning", SeverityError, SeverityWarning, true},
		{"error>=suggestion", SeverityError, SeveritySuggestion, true},
		{"warning>=error", SeverityWarning, SeverityError, false},
		{"warning>=warning", SeverityWarning, SeverityWarning, true},
		{"warning>=suggestion", SeverityWarning, SeveritySuggestion, true},
		{"suggestion>=error", SeveritySuggestion, SeverityError, false},
		{"suggestion>=warning", SeveritySuggestion, SeverityWarning, false},
		{"suggestion>=suggestion", SeveritySuggestion, SeveritySuggestion, true},
		{"suggestion>=invalid", SeveritySuggestion, Severity("bogus"), true},
		{"invalid>=suggestion", Severity("bogus"), SeveritySuggestion, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sev.AtLeast(tt.other); got != tt.want {
				t.Errorf("%q.AtLeast(%q) = %v, want %v", tt.sev, tt.other, got, tt.want)
			}
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Severity
		wantErr bool
	}{
		{"lower error", "error", SeverityError, false},
		{"upper", "WARNING", SeverityWarning, false},
		{"mixed with spaces", "  Suggestion  ", SeveritySuggestion, false},
		{"unknown", "critical", "", true},
		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSeverity(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSeverity(%q) expected error, got nil", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSeverity(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseSeverity(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseSeverityErrorMentionsInput(t *testing.T) {
	_, err := ParseSeverity("zzz")
	if err == nil {
		t.Fatal("expected error")
	}
	if want := "zzz"; !contains(err.Error(), want) {
		t.Errorf("error %q should mention input %q", err.Error(), want)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

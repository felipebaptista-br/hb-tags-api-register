package ids

import (
	"strings"
	"testing"
)

func TestGenerateFormatAndUniqueness(t *testing.T) {
	const n = 5000
	got, err := Generate(n)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(got) != n {
		t.Fatalf("len = %d, want %d", len(got), n)
	}
	seen := make(map[string]struct{}, n)
	for _, id := range got {
		if len(id) != Length {
			t.Fatalf("id %q length = %d", id, len(id))
		}
		if !Valid(id) {
			t.Fatalf("id %q is outside the alphabet", id)
		}
		for _, c := range id {
			if !inAlphabet(byte(c)) {
				t.Fatalf("id %q contains %q", id, c)
			}
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestValidRejectsMalformedIDs(t *testing.T) {
	for _, id := range []string{"", "short", "23456789AI", "0000000000", "23456789AB "} {
		if Valid(id) {
			t.Fatalf("Valid(%q) = true", id)
		}
	}
	if !Valid("23456789AB") {
		t.Fatal("expected a well-formed id to be valid")
	}
}

func TestAlphabetMatchesIDCheck(t *testing.T) {
	if len(Alphabet) != 31 {
		t.Fatalf("alphabet length = %d, want 31", len(Alphabet))
	}
	for i := 0; i < len(Alphabet); i++ {
		id := strings.Repeat("2", Length-1) + string(Alphabet[i])
		if !Valid(id) {
			t.Fatalf("alphabet character %q is rejected by the id check", Alphabet[i])
		}
	}
}

func inAlphabet(c byte) bool {
	for i := 0; i < len(Alphabet); i++ {
		if Alphabet[i] == c {
			return true
		}
	}
	return false
}

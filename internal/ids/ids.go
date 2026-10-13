package ids

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
)

// Alphabet is the tag id alphabet printed on the QR code.
// It matches the database check ^[2-9A-HJKMNP-Z]{10}$.
const Alphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// Length is the fixed size of a tag id.
const Length = 10

var pattern = regexp.MustCompile(`^[2-9A-HJKMNP-Z]{10}$`)

// Valid reports whether id matches the tag primary key format.
func Valid(id string) bool {
	return pattern.MatchString(id)
}

// Generate returns n distinct ids. Each character is chosen with
// crypto/rand.Int bounded by the alphabet size, so the selection has no modulo bias.
func Generate(n int) ([]string, error) {
	if n < 0 {
		return nil, fmt.Errorf("n must be >= 0")
	}
	limit := big.NewInt(int64(len(Alphabet)))
	out := make([]string, 0, n)
	seen := make(map[string]struct{}, n)
	for len(out) < n {
		id, err := newID(limit)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out, nil
}

func newID(limit *big.Int) (string, error) {
	buf := make([]byte, Length)
	for i := 0; i < Length; i++ {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		buf[i] = Alphabet[n.Int64()]
	}
	return string(buf), nil
}

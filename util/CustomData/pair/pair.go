package pair

import (
	"strings"

	gcstr "github.com/PlayerR9/go-commons/strings"
)

// Pair is a pair of values.
type Pair[A, B any] struct {
	// The first value.
	First A

	// The second value.
	Second B
}

// String implements the fmt.Stringer interface.
func (p Pair[A, B]) String() string {
	var builder strings.Builder

	builder.WriteRune('(')
	builder.WriteString(gcstr.GoStringOf(p.First))
	builder.WriteString(", ")
	builder.WriteString(gcstr.GoStringOf(p.Second))
	builder.WriteRune(')')

	return builder.String()
}

// NewPair creates a new pair.
//
// Parameters:
//   - first: The first value.
//   - second: The second value.
//
// Returns:
//   - Pair[A, B]: The new pair.
func NewPair[A, B any](first A, second B) Pair[A, B] {
	p := Pair[A, B]{
		First:  first,
		Second: second,
	}

	return p
}

// ExtractFirsts extracts all the first elements from the given slice of pairs.
//
// Parameters:
//   - pairs: The slice of pairs.
//
// Returns:
//   - []A: The slice of first elements.
func ExtractFirsts[A, B any](pairs []Pair[A, B]) []A {
	if len(pairs) == 0 {
		return nil
	}

	firsts := make([]A, 0, len(pairs))

	for _, pair := range pairs {
		firsts = append(firsts, pair.First)
	}

	return firsts
}

// ExtractSeconds extracts all the second elements from the given slice of pairs.
//
// Parameters:
//   - pairs: The slice of pairs.
//
// Returns:
//   - []B: The slice of second elements.
func ExtractSeconds[A, B any](pairs []Pair[A, B]) []B {
	if len(pairs) == 0 {
		return nil
	}

	seconds := make([]B, 0, len(pairs))

	for _, pair := range pairs {
		seconds = append(seconds, pair.Second)
	}

	return seconds
}

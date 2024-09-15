package c_string

// CStringer is an interface that defines the behavior of a type that can be
// converted to a string representation.
type CStringer interface {
	// CString returns a string representation of the object.
	//
	// Parameters:
	//   - trav: The traversor to use for printing.
	//
	// Returns:
	//   - error: An error if there was a problem generating the string.
	CString(trav *Traversor) error
}

// CStringFunc is a function that generates a formatted string representation of an object.
//
// Parameters:
//   - trav: The traversor to use for printing.
//   - elem: The element to print.
//
// Returns:
//   - error: An error if there was a problem generating the string.
type CStringFunc[T any] func(trav *Traversor, elem T) error

var (
	// ArrayLikeFormat is the default options for an array-like object.
	// [1, 2, 3]
	ArrayLikeFormat FormatConfig = NewFormatter(
		NewDelimiterConfig("[", false, true),
		NewDelimiterConfig("]", false, false),
		NewSeparator(DefaultSeparator, false),
	)
)

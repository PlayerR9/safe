package c_string

import (
	"errors"
	"fmt"
	"strings"

	gcers "github.com/PlayerR9/go-errors"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
)

const (
	// NBSP is the non-breaking space rune.
	NBSP rune = '\u00A0'
)

// ApplyTravFunc applies a function to the printer. Useful for when you want to apply a function
// that does not implement the CStringer interface.
//
// Parameters:
//   - trav: The traversor to use.
//   - elem: The element to apply the function to.
//   - f: The function to apply.
//
// Returns:
//   - error: An error if the function fails.
//
// Errors:
//   - *ErrFinalization: If the finalization of the page fails.
//   - any error returned by the function.
func ApplyTravFunc[T any](trav *Traversor, elem T, f CStringFunc[T]) error {
	err := f(trav, elem)
	if err != nil {
		return err
	}

	return nil
}

// ApplyTravFuncMany applies a function to the printer. Useful for when you want to apply a function
// that does not implement the CStringer interface.
//
// Parameters:
//   - trav: The traversor to use.
//   - f: The function to apply.
//   - elems: The elements to apply the function to.
//
// Returns:
//   - error: An error if the function fails.
//
// Errors:
//   - *ErrFinalization: If the finalization of the page fails.
//   - *Errors.ErrAt: If an error occurs on a specific element.
//   - any error returned by the function.
func ApplyTravFuncMany[T any](trav *Traversor, f CStringFunc[T], elems []T) error {
	if len(elems) == 0 {
		return nil
	}

	for i, elem := range elems {
		err := f(trav, elem)
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" element", err)
		}
	}

	return nil
}

// Traversor is a type that represents a traversor for a formatted string.
type Traversor struct {
	// config is the configuration of the traversor.
	config FormatConfig

	// indentation is the string that is used for indentation
	// on the left side of the traversor.
	indentation []*Unit

	// indentStr is the string that is used for indentation.
	indentStr []*Unit

	// hasIndent is a flag that indicates if the traversor has indentation.
	hasIndent bool

	// source is the buffer of the traversor.
	source *buffer

	// form is the formatter of the traversor.
	form FormatConfig
}

// Cleanup implements the Cleaner interface.
func (trav *Traversor) Cleanup() {
	trav.source = nil
}

// newTraversor creates a new traversor.
//
// Parameters:
//   - config: The configuration of the traversor.
//   - source: The source of the traversor.
//
// Returns:
//   - *Traversor: The new traversor.
func newTraversor(config FormatConfig, source *buffer) *Traversor {
	trav := &Traversor{
		config: config,
		source: source, // shared source
		form:   config,
	}

	indentConfig, ok := config[ConfInd_Idx].(*IndentConfig)
	if !ok {
		panic(fmt.Errorf("invalid configuration type for indentation: %T", config[ConfInd_Idx]))
	}

	if indentConfig == nil {
		trav.hasIndent = false
		trav.indentation = nil
	} else {
		trav.indentation = indentConfig.GetIndentation()
		trav.hasIndent = true
	}

	trav.indentStr = indentConfig.units

	return trav
}

// writeIndent writes the indentation string to the traversor if
// the traversor has indentation and the traversor is at the first
// of the line.
func (trav *Traversor) writeIndent() {
	if trav.hasIndent && trav.source.isFirstOfLine() {
		trav.source.writeIndent(trav.indentStr)
	}
}

// writeRune appends a rune to the current, in-progress line of the traversor.
//
// Parameters:
//   - r: The rune to append.
func (trav *Traversor) writeRune(r rune, style tcell.Style) {
	trav.writeIndent()

	if r == NBSP {
		trav.source.writeString(string(r), style)
	} else {
		trav.source.write(r, style)
	}
}

// writeString appends a string to the current, in-progress line of the traversor.
//
// Parameters:
//   - str: The string to append.
//
// Returns:
//   - error: An error if the string could not be appended.
func (trav *Traversor) writeString(str string, style tcell.Style) error {
	trav.writeIndent()

	if str == "" {
		return nil
	}

	n := checkString(str)
	if n != -1 {
		return gcers.NewErrAt(humanize.Ordinal(n+1)+" rune", errors.New("not proper UTF-8 encoding"))
	}

	trav.source.writeString(str, style)

	return nil
}

// writeLine writes a line to the traversor. If there is any in-progress line,
// then the line is appended to the line before accepting it. Otherwise, a new line
// with the line is added to the source.
//
// Parameters:
//   - line: The line to write.
//
// Returns:
//   - error: An error if the line could not be written.
//
// Behaviors:
//   - If line is empty, then an empty line is added to the source.
func (trav *Traversor) writeLine(line string, style tcell.Style) error {
	trav.source.acceptLine() // Accept the current line if any.

	trav.writeIndent()

	if line == "" {
		trav.source.writeEmptyLine()
	} else {
		n := checkString(line)
		if n != -1 {
			return gcers.NewErrAt(humanize.Ordinal(n+1)+" rune", errors.New("not proper UTF-8 encoding"))
		}

		trav.source.writeString(line, style)
	}

	trav.source.acceptLine() // Accept the line.

	return nil
}

// AppendRune appends a rune to the half-line of the traversor.
//
// Parameters:
//   - r: The rune to append.
//
// Behaviors:
//   - If the half-line is nil, then a new half-line is created.
func (trav *Traversor) AppendRune(r rune, style tcell.Style) {
	if trav.source != nil {
		trav.writeRune(r, style)
	}
}

// AppendString appends a string to the half-line of the traversor.
//
// Parameters:
//   - str: The string to append.
//
// Returns:
//   - error: An error of type *Errors.ErrInvalidRuneAt if there is an invalid rune
//     in the string.
//
// Behaviors:
//   - IF str is empty: nothing is done.
func (trav *Traversor) AppendString(str string, style tcell.Style) error {
	if trav.source == nil {
		return nil
	}

	return trav.writeString(str, style)
}

// AppendStrings appends multiple strings to the half-line of the traversor.
//
// Parameters:
//   - strs: The strings to append.
//
// Returns:
//   - error: An error of type *Errors.ErrAt if there is an error appending a string.
//
// Behaviors:
//   - This is equivalent to calling AppendString for each string in strs but more efficient.
func (trav *Traversor) AppendStrings(strs []string, style tcell.Style) error {
	if trav.source == nil || len(strs) == 0 {
		return nil
	}

	for i, str := range strs {
		err := trav.writeString(str, style)
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" string", err)
		}
	}

	return nil
}

// AppendJoinedString appends a joined string to the half-line of the traversor.
//
// Parameters:
//   - sep: The separator to use.
//   - fields: The fields to join.
//
// Returns:
//   - error: An error of type *Errors.ErrInvalidRuneAt if there is an invalid rune
//     in the string.
//
// Behaviors:
//   - This is equivalent to calling AppendString(strings.Join(fields, sep)).
func (trav *Traversor) AppendJoinedString(style tcell.Style, sep string, fields ...string) error {
	if trav.source == nil || len(fields) == 0 {
		return nil
	}

	str := strings.Join(fields, sep)

	err := trav.writeString(str, style)
	if err != nil {
		return err
	}

	return nil
}

// AcceptWord is a function that, if there is any in-progress word, then said word is added
// to the source.
func (trav *Traversor) AcceptWord() {
	if trav.source == nil {
		return
	}

	trav.source.acceptWord()
}

// AcceptLine is a function that accepts the current line of the traversor.
//
// Behaviors:
//   - This also accepts the current word if any.
func (trav *Traversor) AcceptLine() {
	if trav.source == nil {
		return
	}

	trav.source.acceptLine()
}

// AddLine adds a line to the traversor. If there is any in-progress line, then the line is
// appended to the line before accepting it. Otherwise, a new line with the line is added to
// the source.
//
// Parameters:
//   - line: The line to add.
//
// Returns:
//   - error: An error of type *Errors.ErrAt if there is an error adding the line.
//
// Behaviors:
//   - If line is empty, then an empty line is added to the source.
func (trav *Traversor) AddLine(line string, style tcell.Style) error {
	if trav.source == nil {
		return nil
	}

	return trav.writeLine(line, style)
}

// AddLines adds multiple lines to the traversor in a more efficient way than
// adding each line individually.
//
// Parameters:
//   - lines: The lines to add.
//
// Returns:
//   - error: An error of type *Errors.ErrAt if there is an error adding a line.
//
// Behaviors:
//   - If there are no lines, then nothing is done.
func (trav *Traversor) AddLines(lines []string, style tcell.Style) error {
	if trav.source == nil || len(lines) == 0 {
		return nil
	}

	for i, line := range lines {
		err := trav.writeLine(line, style)
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" line", err)
		}
	}

	return nil
}

// AddJoinedLine adds a joined line to the traversor. This is a more efficient way to do
// the same as AddLine(strings.Join(fields, sep)).
//
// Parameters:
//   - sep: The separator to use.
//   - fields: The fields to join.
//
// Returns:
//   - error: An error of type *Errors.ErrInvalidRuneAt if there is an invalid rune
//     in the line.
//
// Behaviors:
//   - If fields is empty, then nothing is done.
func (trav *Traversor) AddJoinedLine(style tcell.Style, sep string, fields ...string) error {
	if trav.source == nil || len(fields) == 0 {
		return nil
	}

	str := strings.Join(fields, sep)

	err := trav.writeLine(str, style)
	if err != nil {
		return err
	}

	return nil
}

// EmptyLine adds an empty line to the traversor. This is a more efficient way to do
// the same as AddLine("") or AddLines([]string{""}).
//
// Behaviors:
//   - If the half-line is not empty, then the half-line is added to the source
//     (half-line is reset) and an empty line is added to the source.
func (trav *Traversor) EmptyLine() {
	if trav.source == nil {
		return
	}

	trav.source.acceptLine() // Accept the current line if any.

	trav.writeIndent()

	trav.source.acceptLine() // Accept the line.
}

// Write implements the io.Writer interface for the traversor.
func (trav *Traversor) Write(p []byte) (int, error) {
	if trav.source == nil {
		return 0, nil
	}

	var style tcell.Style

	config, ok := trav.form[ConfStyle_Idx].(*StyleConfig)
	if !ok {
		style = tcell.StyleDefault
	} else {
		style = config.defaultStyle
	}

	trav.source.writeString(string(p), style)

	return len(p), nil
}

// Print is a function that writes to the traversor using the fmt.Fprint function.
//
// Parameters:
//   - a: The arguments to write.
func (trav *Traversor) Print(a ...interface{}) error {
	if trav.source == nil {
		return nil
	}

	_, err := fmt.Fprint(trav, a...)
	return err
}

// Printf is a function that writes to the traversor using the fmt.Fprintf function.
//
// Parameters:
//   - format: The format string.
//   - a: The arguments to write.
func (trav *Traversor) Printf(format string, a ...interface{}) error {
	if trav.source == nil {
		return nil
	}

	_, err := fmt.Fprintf(trav, format, a...)
	return err
}

// Println is a function that writes to the traversor using the fmt.Fprintln function.
//
// Parameters:
//   - a: The arguments to write.
func (trav *Traversor) Println(a ...interface{}) error {
	if trav.source == nil {
		return nil
	}

	_, err := fmt.Fprintln(trav, a...)
	return err
}

// ConfigOption is a type that represents a configuration option for a formatter.
type ConfigOption func(FormatConfig)

// WithIncreasedIndent is a function that increases the indentation level of the formatter
// by one.
//
// Returns:
//   - ConfigOption: The configuration option.
func WithIncreasedIndent() ConfigOption {
	return func(f FormatConfig) {
		config, ok := f[ConfInd_Idx].(*IndentConfig)
		if !ok {
			panic(fmt.Errorf("invalid configuration type for indentation: %T", f[ConfInd_Idx]))
		}

		if config != nil {
			config.level++
		}
	}
}

// WithDecreasedIndent is a function that decreases the indentation level of the formatter
// by one.
//
// Returns:
//   - ConfigOption: The configuration option.
//
// Behaviors:
//   - If the indentation level is already 0, it is not decreased.
func WithDecreasedIndent() ConfigOption {
	return func(f FormatConfig) {
		config, ok := f[ConfInd_Idx].(*IndentConfig)
		if !ok {
			panic(fmt.Errorf("invalid configuration type for indentation: %T", f[ConfInd_Idx]))
		}

		if config != nil && config.level > 0 {
			config.level--
		}
	}
}

// WithModifiedIndent is a function that modifies the indentation level of the formatter
// by a specified amount relative to the current indentation level.
//
// Parameters:
//   - by: The amount by which to modify the indentation level.
//
// Returns:
//   - ConfigOption: The configuration option.
//
// Behaviors:
//   - Negative values will decrease the indentation level while positive values will
//     increase it. If the value is 0, then nothing is done and when the indentation level
//     is 0, it is not decreased.
func WithModifiedIndent(by int) ConfigOption {
	if by == 0 {
		return func(f FormatConfig) {}
	} else {
		return func(f FormatConfig) {
			config, ok := f[ConfInd_Idx].(*IndentConfig)
			if !ok {
				panic(fmt.Errorf("invalid configuration type for indentation: %T", f[ConfInd_Idx]))
			}

			if config == nil {
				return
			}

			config.level += by
			if config.level < 0 {
				config.level = 0
			}
		}
	}
}

// GetConfig is a method that returns a copy of the configuration of the traversor.
//
// Parameters:
//   - options: The options to apply to the configuration.
//
// Returns:
//   - FormatConfig: A copy of the configuration of the traversor.
func (trav *Traversor) GetConfig(options ...ConfigOption) FormatConfig {
	var configCopy FormatConfig

	// FIXME: This is a hack.
	/* for i := 0; i < 4; i++ {
		configCopy[i] = trav.config[i].Copy()
	} */

	for _, option := range options {
		option(configCopy)
	}

	return configCopy
}

//////////////////////////////////////////////////////////////

/*
// GetIndent returns the indentation string of the traversor.
//
// Returns:
//   - string: The indentation string of the traversor.
func (trav *Traversor) GetIndent() string {
	if trav.indent == nil {
		return ""
	} else {
		return trav.indentStr
	}
}

// ApplyIndent applies the indentation configuration to a specified string.
//
// Parameters:
//   - str: The string to apply the indentation to.
//
// Returns:
//   - string: The string with the indentation applied.
func (trav *Traversor) ApplyIndent(isFirstLine bool, str string) string {
	if trav.indent == nil || !trav.source.isFirstOfLine() {
		return str
	}

	var builder strings.Builder

	builder.WriteString(trav.indentStr)
	builder.WriteString(str)

	return builder.String()
}
*/

/*
// AddMultiline adds a multiline to the traversor. But first, it accepts any in-progress
// half-line.
//
// Parameters:
//   - mlt: The multiline to add.
//
// Behaviors:
//   - If the multiline is nil, then nothing is done.
func (trav *Traversor) AddMultiline(mlt *cb.MultiLineText) {
	if mlt == nil {
		return
	}

	trav.AcceptHalfLine()
	trav.source.addLine(mlt)
}
*/

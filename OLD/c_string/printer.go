package c_string

import (
	gcers "github.com/PlayerR9/errors"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell"
)

// Printer is a type that represents a formatted string.
type Printer struct {
	// buffer is the buffer of the document.
	buff *buffer

	// formatter is the formatter of the document.
	formatter FormatConfig
}

// NewPrinter creates a new printer.
//
// Parameters:
//   - form: The formatter to use.
//
// Returns:
//   - *Printer: The new printer.
//
// Behaviors:
//   - If the formatter is nil, the function uses the formatter with nil values.
func NewPrinter(form FormatConfig) *Printer {
	return &Printer{
		buff:      newBuffer(),
		formatter: form,
	}
}

// NewPrinterFromConfig creates a new printer from a configuration.
//
// Parameters:
//   - opts: The configuration to use.
//
// Returns:
//   - *Printer: The new printer.
//
// Behaviors:
//   - If the configuration is nil, the function uses the default configuration.
//   - Panics if an invalid configuration type is given (i.e., not IndentConfig, DelimiterConfig,
//     or SeparatorConfig).
func NewPrinterFromConfig(opts ...any) *Printer {
	return &Printer{
		buff:      newBuffer(),
		formatter: NewFormatter(opts...),
	}
}

// GetTraversor returns a traversor for the printer.
//
// Returns:
//   - *Traversor: The traversor for the printer.
func (p *Printer) GetTraversor() *Traversor {
	return newTraversor(p.formatter, p.buff)
}

// Apply applies a format to a stringer.
//
// Parameters:
//   - p: The printer to use.
//   - elem: The element to format.
//
// Returns:
//   - error: An error if the formatting fails.
//
// Errors:
//   - *ErrInvalidParameter: If the printer is nil.
//   - *ErrFinalization: If the finalization of the page fails.
//   - any error returned by the element's CString method.
//
// Behaviors:
//   - If the formatter is nil, the function uses the nil formatter.
//   - If the element is nil, the function does nothing.
func Apply[T CStringer](p *Printer, elem T) error {
	if p == nil {
		return gcers.NewErrNilParameter("p")
	}

	trav := newTraversor(p.formatter, p.buff)

	err := elem.CString(trav)
	if err != nil {
		return err
	}

	return nil
}

// ApplyMany applies a format to a stringer.
//
// Parameters:
//   - p: The printer to use.
//   - elems: The elements to format.
//
// Returns:
//   - error: An error if the formatting fails.
//
// Errors:
//   - *ErrInvalidParameter: If the printer is nil.
//   - *ErrFinalization: If the finalization of the page fails.
//   - *Errors.ErrAt: If an error occurs on a specific element.
//
// Behaviors:
//   - If the formatter is nil, the function uses the nil formatter.
//   - If an element is nil, the function skips the element.
//   - If all elements are nil, the function does nothing.
func ApplyMany[T CStringer](p *Printer, elems []T) error {
	if len(elems) == 0 {
		return nil
	}

	if p == nil {
		return gcers.NewErrNilParameter("p")
	}

	for i, elem := range elems {
		err := elem.CString(newTraversor(p.formatter, p.buff))
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" element", err)
		}
	}

	return nil
}

// ApplyFunc applies a format function to the printer.
//
// Parameters:
//   - p: The printer to use.
//   - elem: The element to apply the function to.
//   - f: The function to apply.
//
// Returns:
//   - error: An error if the function fails.
//
// Errors:
//   - *ErrInvalidParameter: If the printer is nil.
//   - any error returned by the function.
func ApplyFunc[T any](p *Printer, elem T, f CStringFunc[T]) error {
	if p == nil {
		return gcers.NewErrNilParameter("p")
	}

	trav := newTraversor(p.formatter, p.buff)

	err := f(trav, elem)
	if err != nil {
		return err
	}

	return nil
}

// ApplyFuncMany applies a format function to the printer.
//
// Parameters:
//   - p: The printer to use.
//   - f: The function to apply.
//   - elems: The elements to apply the function to.
//
// Returns:
//   - error: An error if the function fails.
//
// Errors:
//   - *ErrInvalidParameter: If the printer is nil.
//   - *Errors.ErrAt: If an error occurs on a specific element.
//   - any error returned by the function.
func ApplyFuncMany[T any](p *Printer, f CStringFunc[T], elems []T) error {
	if len(elems) == 0 {
		return nil
	}

	if p == nil {
		return gcers.NewErrNilParameter("p")
	}

	for i, elem := range elems {
		err := f(newTraversor(p.formatter, p.buff), elem)
		if err != nil {
			return gcers.NewErrAt(humanize.Ordinal(i+1)+" element", err)
		}
	}

	return nil
}

// GetPages returns the pages of the printer.
//
// Returns:
//   - [][][][]string: The pages of the printer.
func (p *Printer) GetPages() [][][][][]*Unit {
	p.buff.finalize()

	pages := p.buff.pages

	// Reset the buffer
	p.buff = newBuffer()

	allStrings := make([][][][][]*Unit, 0, len(pages))

	for _, page := range pages {
		sectionLines := make([][][][]*Unit, 0)

		for _, section := range page {
			sectionLines = append(sectionLines, section.getLines())
		}

		allStrings = append(allStrings, sectionLines)
	}

	return allStrings
}

// Cleanup implements the Cleaner interface.
func (p *Printer) Cleanup() {
	p.buff.Cleanup()

	p.buff = nil
}

// Printc prints a character.
//
// Parameters:
//   - form: The formatter to use.
//   - c: The character to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted character.
func Printc(form FormatConfig, style tcell.Style, c rune) [][][][][]*Unit {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	trav.AppendRune(c, style)

	return p.GetPages()
}

// Print prints a string.
//
// Parameters:
//   - form: The formatter to use.
//   - strs: The strings to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Print(form FormatConfig, style tcell.Style, strs ...string) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	var err error

	switch len(strs) {
	case 0:
		// Do nothing
	case 1:
		err = trav.AppendString(strs[0], style)
	default:
		err = trav.AppendStrings(strs, style)
	}

	if err != nil {
		return nil, err
	}

	// apply

	return p.GetPages(), nil
}

// Printj prints a joined string.
//
// Parameters:
//   - form: The formatter to use.
//   - sep: The separator to use.
//   - strs: The strings to join.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Printj(form FormatConfig, style tcell.Style, sep string, strs ...string) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	err := trav.AppendJoinedString(style, sep, strs...)
	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

// Fprint prints a formatted string.
//
// Parameters:
//   - form: The formatter to use.
//   - a: The elements to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Fprint(form FormatConfig, style tcell.Style, a ...interface{}) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	err := trav.Print()
	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

// Fprintf prints a formatted string.
//
// Parameters:
//   - form: The formatter to use.
//   - format: The format string.
//   - a: The elements to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Fprintf(form FormatConfig, style tcell.Style, format string, a ...interface{}) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	err := trav.Printf(format, a...)
	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

// Println prints a string with a newline.
//
// Parameters:
//   - form: The formatter to use.
//   - lines: The lines to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Println(form FormatConfig, style tcell.Style, lines ...string) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	var err error

	switch len(lines) {
	case 0:
		trav.EmptyLine()
	case 1:
		err = trav.AddLine(lines[0], style)
	default:
		err = trav.AddLines(lines, style)
	}

	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

// Printjln prints a joined string with a newline.
//
// Parameters:
//   - form: The formatter to use.
//   - sep: The separator to use.
//   - lines: The lines to join.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Printjln(form FormatConfig, style tcell.Style, sep string, lines ...string) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	err := trav.AddJoinedLine(style, sep, lines...)
	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

// Fprintln prints a formatted string with a newline.
//
// Parameters:
//   - form: The formatter to use.
//   - a: The elements to print.
//
// Returns:
//   - [][][][]string: The pages of the formatted strings.
//   - error: An error if the printing fails.
func Fprintln(form FormatConfig, style tcell.Style, a ...interface{}) ([][][][][]*Unit, error) {
	p := NewPrinter(form)

	trav := newTraversor(form, p.buff)

	err := trav.Println(a...)
	if err != nil {
		return nil, err
	}

	return p.GetPages(), nil
}

//////////////////////////////////////////////////////////////
/*
const (
	// Hellip is the ellipsis character.
	Hellip string = "..."

	// HellipLen is the length of the ellipsis character.
	HellipLen int = len(Hellip)

	// MarginLeft is the left margin of the content box.
	MarginLeft int = 1
)


// addLine is a private function that adds a line to the formatted string.
//
// Parameters:
//   - mlt: The line to add.
func (p *printerSource) addLine(mlt *cb.MultiLineText) {
	if mlt == nil {
		return
	}

	p.lines = append(p.lines, mlt)
}

// GetLines returns the lines of the formatted string.
//
// Returns:
//   - []*MultiLineText: The lines of the formatted string.
func (p *printerSource) GetLines() []*cb.MultiLineText {
	return p.lines
}

/*
func (p *printerSource) Boxed(width, height int) ([]string, error) {
	p.fix()

	all_fields := p.getAllFields()

	fss := make([]*printerSource, 0, len(all_fields))

	for _, fields := range all_fields {
		p := &printerSource{
			lines: fields,
		}

		fss = append(fss, p)
	}

	lines := make([]string, 0)

	for _, p := range fss {
		ts, err := p.generateContentBox(width, height)
		if err != nil {
			return nil, err
		}

		leftLimit, ok := ts.GetFurthestRightEdge()
		if !ok {
			panic("could not get furthest right edge")
		}

		for _, line := range ts.GetLines() {
			fitted, err := sext.FitString(line.String(), leftLimit)
			if err != nil {
				return nil, err
			}

			lines = append(lines, fitted)
		}
	}

	return lines, nil
}


func (p *printerSource) fix() {
	// 1. Fix newline boundaries
	newLines := make([]string, 0)

	for _, line := range p.lines {
		newFields := strings.Split(line, "\n")

		newLines = append(newLines, newFields...)
	}

	p.lines = newLines
}

// Must call Fix() before calling this function.
func (p *printerSource) getAllFields() [][]string {
	// TO DO: Handle special WHITESPACE characters

	fieldList := make([][]string, 0)

	for _, content := range p.lines {
		fields := strings.Fields(content)

		if len(fields) != 0 {
			fieldList = append(fieldList, fields)
		}
	}

	return fieldList
}
*/

/*
// GetDocument returns the content of the FieldSplitter as a Document.
//
// Returns:
//   - *Document: The content of the FieldSplitter.
func (p *FieldSplitter) GetDocument() *FieldSplitter {
	return p.content
}


// Build is a function that builds the document.
//
// Returns:
//   - *tld.Document: The built document.
func (p *FieldSplitter) Build() *tld.Document {
	doc := tld.NewDocument()

	for _, page := range p.content.pages {
		iter := page.Iterator()

		for {
			section, err := iter.Consume()
			if err != nil {
				break
			}
		}
	}

	return doc
}
*/

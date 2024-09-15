package c_string

import (
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell"
)

// checkString is a private function that checks a string for invalid runes.
//
// Parameters:
//   - str: The string to check.
//
// Returns:
//   - int: The index of the invalid rune. -1 if no invalid runes are found.
func checkString(str string) int {
	for j := 0; len(str) > 0; j++ {
		char, size := utf8.DecodeRuneInString(str)
		str = str[size:]

		if char == utf8.RuneError {
			return j
		}

		if char == '\r' && size != 0 {
			nextRune, size := utf8.DecodeRuneInString(str)

			if nextRune == '\n' {
				str = str[size:]
			}
		}
	}

	return -1
}

// unitBuffer is a type that represents a buffer of units.
type unitBuffer struct {
	// units are the units of the buffer.
	units []*Unit
}

// Cleanup implements the Cleanup interface method.
func (ub *unitBuffer) Cleanup() {
	for i := 0; i < len(ub.units); i++ {
		ub.units[i] = nil
	}

	ub.units = nil
}

// Len is a function that returns the number of units in the buffer.
//
// Returns:
//   - int: The number of units in the buffer.
func (ub *unitBuffer) Len() int {
	return len(ub.units)
}

// removeLast is a function that removes the last unit from the buffer.
//
// Returns:
//   - bool: True if a unit was removed. False otherwise.
func (ub *unitBuffer) removeLast() bool {
	for i := len(ub.units) - 1; i >= 0; i-- {
		ok := ub.units[i].removeLastRune()
		if ok {
			return true
		}
	}

	return false
}

// getUnits is a function that returns the units of the buffer.
//
// Returns:
//   - []*Unit: The units of the buffer.
func (ub *unitBuffer) getUnits() []*Unit {
	return ub.units
}

// WriteString is a function that adds a string to the buffer.
//
// Parameters:
//   - str: The string to add.
//   - style: The style of the string.
func (ub *unitBuffer) WriteString(str string, style tcell.Style) {
	newUnit := NewUnit(str, style)

	if len(ub.units) == 0 {
		ub.units = append(ub.units, newUnit)
	} else {
		ok := ub.units[len(ub.units)-1].Merge(newUnit)
		if !ok {
			ub.units = append(ub.units, newUnit)
		}
	}
}

// sectionBuilder is a type that represents a section of a page.
type sectionBuilder struct {
	// buff is the string buff for the section.
	buff *unitBuffer

	// lines are the lines in the section.
	lines [][][]*Unit

	// lastLine is the last line of the section.
	lastLine int

	// mu is the mutex for the builder.
	mu sync.RWMutex
}

// Cleanup implements the Cleanup interface method.
func (sb *sectionBuilder) Cleanup() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	for i := 0; i < len(sb.lines); i++ {
		sb.lines[i] = nil
	}

	sb.lines = nil

	sb.buff.Cleanup()
}

// newSectionBuilder creates a new section.
//
// Returns:
//   - *Section: The new section.
func newSectionBuilder() *sectionBuilder {
	return &sectionBuilder{
		lines:    [][][]*Unit{{}},
		lastLine: 0,
	}
}

// removeOne is a function that removes the last character from the section.
//
// Returns:
//   - bool: True if a character was removed. False otherwise.
func (sb *sectionBuilder) removeOne() bool {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	ok := sb.buff.removeLast()
	if ok {
		return true
	}

	for i := sb.lastLine; i >= 0; i-- {
		words := sb.lines[i]

		for j := len(words) - 1; j >= 0; j-- {
			if len(words[j]) > 0 {
				words[j] = words[j][:len(words[j])-1]
				return true
			}
		}
	}

	return false
}

// getLines is a function that returns the words of the section.
//
// Returns:
//   - [][]string: The words of the section.
func (sb *sectionBuilder) getLines() [][][]*Unit {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.lines
}

// isFirstOfLine is a function that returns true if the current position is the first
// position of a line.
//
// Returns:
//   - bool: True if the current position is the first position of a line.
func (sb *sectionBuilder) isFirstOfLine() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.buff.Len() == 0 && len(sb.lines[sb.lastLine]) == 0
}

// accept is a function that accepts the current word and
// creates a new line.
func (sb *sectionBuilder) accept() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.buff.Len() > 0 {
		sb.lines[sb.lastLine] = append(sb.lines[sb.lastLine], sb.buff.getUnits())
		sb.buff.Cleanup()
	}

	sb.lines = append(sb.lines, [][]*Unit{})
	sb.lastLine++
}

// acceptWord is a function that accepts the current in-progress word
// and resets the builder.
func (sb *sectionBuilder) acceptWord() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.buff.Len() == 0 {
		return
	}

	sb.lines[sb.lastLine] = append(sb.lines[sb.lastLine], sb.buff.getUnits())
	sb.buff.Cleanup()
}

// writeString adds a string to the current, in-progress word.
//
// Parameters:
//   - str: The string to write.
//   - style: The style of the string.
func (sb *sectionBuilder) writeString(str string, style tcell.Style) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.buff.WriteString(str, style)
}

// writeUnits adds a list of units to the current, in-progress word.
//
// Parameters:
//   - units: The units to write.
func (sb *sectionBuilder) writeUnits(units []*Unit) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	for i := 0; i < len(units); i++ {
		sb.buff.units = append(sb.buff.units, units[i])
	}
}

// buffer is a type that represents a buffer of a document.
type buffer struct {
	// pages are the pages of the buffer.
	pages [][]*sectionBuilder

	// buff is the in-progress section of the buffer.
	buff *sectionBuilder

	// lastPage is the last page of the buffer.
	lastPage int
}

// Cleanup implements the Cleanup interface method.
func (b *buffer) Cleanup() {
	// pages are the pages of the buffer.
	for i := 0; i < len(b.pages); i++ {
		for j := 0; j < len(b.pages[i]); j++ {
			b.pages[i][j].Cleanup()
			b.pages[i][j] = nil
		}

		b.pages[i] = nil
	}

	b.pages = nil

	b.buff.Cleanup()
	b.buff = nil
}

// newBuffer creates a new buffer.
//
// Returns:
//   - *buffer: The new buffer.
func newBuffer() *buffer {
	return &buffer{
		pages:    [][]*sectionBuilder{{}},
		buff:     nil,
		lastPage: 0,
	}
}

// isFirstOfLine is a private function that returns true if the current position is the first
// position of a line.
//
// Returns:
//   - bool: True if the current position is the first position of a line.
func (b *buffer) isFirstOfLine() bool {
	return b.buff == nil || b.buff.isFirstOfLine()
}

// writeIndent is a private function that writes the indentation to the formatted string.
//
// Parameters:
//   - str: The string to write.
//   - style: The style of the string.
func (b *buffer) writeIndent(units []*Unit) {
	if len(units) == 0 {
		return
	}

	if b.buff == nil {
		b.buff = newSectionBuilder()
	}

	b.buff.writeUnits(units)
}

// Accept is a function that accepts the current in-progress buffer
// by converting it to the specified section type. Lastly, the section
// is added to the page.
//
// Parameters:
//   - sectionType: The section type to convert the buffer to.
//
// Behaviors:
//   - Even when the buffer is empty, the section is still added to the page.
//     To avoid this, use the Finalize function.
func (b *buffer) accept() {
	if b.buff != nil {
		b.buff.acceptWord()
	}

	b.pages[b.lastPage] = append(b.pages[b.lastPage], b.buff)

	b.buff = nil
}

// write is a private function that appends a rune to the buffer
// while dealing with special characters.
//
// Parameters:
//   - char: The rune to append.
func (b *buffer) write(char rune, style tcell.Style) {
	switch char {
	case '\t':
		// Tab : Add spaces until the next tab stop
		if b.buff == nil {
			b.buff = newSectionBuilder()
		} else {
			b.buff.acceptWord()
		}

		b.buff.writeString(string(char), style)

		b.buff.acceptWord()
	case '\v':
		// vertical tab : Add vertical tabulation

		// Do nothing
	case '\r', '\n', '\u0085':
		// carriage return : Move to the start of the line (alone)
		// or move to the start of the line and down (with line feed)
		// line feed : Add a new line or move to the left edge and down

		b.accept()
	case '\f':
		// form feed : Go to the next page
		b.accept()

		b.lastPage++
		b.pages = append(b.pages, []*sectionBuilder{})
	case ' ':
		// Space
		if b.buff != nil {
			b.buff.acceptWord()
		}
	case '\u0000', '\a':
		// null : Ignore this character
		// Bell : Ignore this character
	case '\b':
		// backspace : Remove the last character
		if b.buff != nil {
			ok := b.buff.removeOne()
			if ok {
				return
			}
		}

		for i := b.lastPage; i >= 0; i-- {
			sections := b.pages[i]

			for j := len(sections) - 1; j >= 0; j-- {
				section := sections[j]

				ok := section.removeOne()
				if ok {
					return
				}
			}
		}
	case '\u001A':
		// Control-Z : End of file for Windows text-mode file i/o
		b.finalize()
	case '\u001B':
		// escape : Introduce an escape sequence (next character)
		// Do nothing
	default:
		// NBSP : Non-breaking space
		// any other normal character
		if b.buff == nil {
			b.buff = newSectionBuilder()
		}

		if char == NBSP {
			// Non-breaking space
			b.buff.writeString(" ", style)
		} else {
			b.buff.writeString(string(char), style)
		}
	}
}

// writeRune is a private function that appends a rune to the buffer
// without checking for special characters.
//
// Parameters:
//   - r: The rune to append.
//   - style: The style of the rune.
func (b *buffer) writeString(str string, style tcell.Style) {
	if b.buff == nil {
		b.buff = newSectionBuilder()
	}

	b.buff.writeString(str, style)
}

// acceptWord is a private function that accepts the current word of the formatted string.
func (b *buffer) acceptWord() {
	if b.buff != nil {
		b.buff.acceptWord()
	}
}

// acceptLine is a private function that accepts the current line of the formatted string.
func (b *buffer) acceptLine() {
	if b.buff != nil {
		b.buff.accept()
	}
}

// writeEmptyLine is a private function that accepts the current line
// regardless of the whether the line is empty or not.
func (b *buffer) writeEmptyLine() {
	if b.buff == nil {
		b.buff = newSectionBuilder()
	}

	b.buff.accept()
}

// finalize is a private function that finalizes the buffer.
func (b *buffer) finalize() {
	if b.buff == nil {
		return
	}

	b.buff.acceptWord()

	b.pages[b.lastPage] = append(b.pages[b.lastPage], b.buff)

	b.buff = nil
}

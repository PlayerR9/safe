package c_string

import (
	"unicode/utf8"

	"github.com/gdamore/tcell"
)

// Unit is a unit of content that can be displayed.
type Unit struct {
	// Content is the content of the unit.
	Content string

	// Style is the style of the unit.
	Style tcell.Style
}

// Copy is a method that creates a copy of the unit.
//
// Returns:
//   - *Unit: The copy of the unit.
func (u *Unit) Copy() *Unit {
	return &Unit{
		Content: u.Content,
		Style:   u.Style,
	}
}

// NewUnit is a function that creates a new unit.
//
// Parameters:
//   - content: The content of the new unit.
//   - style: The style of the new unit.
//
// Returns:
//   - *Unit: The new unit.
func NewUnit(content string, style tcell.Style) *Unit {
	return &Unit{
		Content: content,
		Style:   style,
	}
}

// Merge is a method that merges the content of the unit with another unit
// if the styles are the same.
//
// Parameters:
//   - other: The other unit to merge with.
//
// Returns:
//   - bool: True if the merge was successful, false otherwise.
func (u *Unit) Merge(other *Unit) bool {
	if other == nil {
		return true
	}

	if u.Style != other.Style {
		return false
	}

	u.Content += other.Content

	return true
}

// removeLastRune removes the last rune from the content of the unit.
//
// Returns:
//   - bool: True if the last rune was removed, false otherwise.
func (u *Unit) removeLastRune() bool {
	if len(u.Content) == 0 {
		return false
	}

	_, size := utf8.DecodeLastRuneInString(u.Content)
	u.Content = u.Content[:len(u.Content)-size]

	return true
}

func ReduceUnitSequence(units []*Unit) []*Unit {
	if len(units) == 0 {
		return units
	}

	// 1. Remove any nil units
	top := 0

	for i := 0; i < len(units); i++ {
		if units[i] != nil {
			units[top] = units[i]
			top++
		}
	}

	units = units[:top]

	// 2. Merge units with the same style
	newSequence := make([]*Unit, 0)

	var currentUnit *Unit = nil

	i := 0
	j := i + 1

	for j < len(units) {
		currentUnit = units[i]

		ok := currentUnit.Merge(units[j])
		if ok {
			j++
		} else {
			newSequence = append(newSequence, currentUnit)
			i = j
			j = i + 1
			currentUnit = nil
		}
	}

	if currentUnit != nil {
		newSequence = append(newSequence, currentUnit)
	}

	return newSequence
}

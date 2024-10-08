package dt_table

import (
	"github.com/gdamore/tcell"

	gcc "github.com/PlayerR9/go-commons/CustomData/common"
)

// DtCell represents a cell in a data table.
type DtCell gcc.Pair[rune, tcell.Style]

// NewDtCell creates a new DtCell with the given content and style.
//
// Parameters:
//   - content: The content of the cell.
//   - style: The style of the cell.
//
// Returns:
//   - *DtCell: A pointer to the new DtCell.
func NewDtCell(content rune, style tcell.Style) *DtCell {
	return &DtCell{
		First:  content,
		Second: style,
	}
}

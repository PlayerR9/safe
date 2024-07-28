package DtTable

import (
	"errors"
	"sync"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

// DtRow represents a row in a data table.
// It is safe for concurrent use.
type DtRow struct {
	// cells is a slice of cells in the row.
	cells []*DtCell

	// width represents the width of the row.
	width int

	// mu is a mutex that protects the row.
	mu sync.RWMutex
}

// NewDtRow creates a new DtRow with the given width all initialized to nil.
//
// Parameters:
//   - width: The width of the row.
//
// Returns:
//   - *DtRow: A pointer to the new DtRow.
//   - error: An error of type *uc.ErrInvalidParameter if the width
//     is less than 0.
func NewDtRow(width int) (*DtRow, error) {
	if width < 0 {
		return nil, uc.NewErrInvalidParameter(
			"width",
			errors.New("value must be non-negative"),
		)
	}

	return &DtRow{
		cells: make([]*DtCell, width),
		width: width,
	}, nil
}

// GetWidth returns the width of the row.
//
// Returns:
//   - int: The width of the row.
func (r *DtRow) GetWidth() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.width
}

// SetCell sets the cell at the given index.
//
// Parameters:
//   - cell: The cell to set.
//   - x: The x-coordinate of the cell.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if the index
//     is out of bounds.
func (r *DtRow) SetCell(cell *DtCell, x int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if x < 0 || x >= r.width {
		return uc.NewErrInvalidParameter(
			"x",
			uc.NewErrOutOfBounds(x, 0, r.width),
		)
	}

	r.cells[x] = cell

	return nil
}

// SetCells sets the cells at the given index.
//
// Parameters:
//   - cells: The cells to set.
//   - from: The index to start setting the cells.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if cells
//     cannot be set at the given index.
func (r *DtRow) SetCells(cells []*DtCell, from int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if from < 0 || from+len(cells) > r.width {
		return uc.NewErrInvalidParameter(
			"from",
			uc.NewErrOutOfBounds(from, 0, r.width-len(cells)),
		)
	}

	for i, cell := range cells {
		r.cells[from+i] = cell
	}

	return nil
}

// Append appends the given cells to the row.
//
// Parameters:
//   - cells: The cells to append.
func (r *DtRow) Append(cells ...*DtCell) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cells = append(r.cells, cells...)
	r.width += len(cells)
}

// GetCellAt returns the cell at the given index.
// If the index is out of bounds, it returns nil.
//
// Parameters:
//   - x: The x-coordinate of the cell.
//
// Returns:
//   - *DtCell: The cell at the given index.
func (r *DtRow) GetCellAt(x int) *DtCell {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if x < 0 || x >= r.width {
		return nil
	}

	return r.cells[x]
}

// Resize resizes the row to the given width.
//
// Parameters:
//   - newWidth: The new width of the row.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if the new width
//     is less than 0.
func (r *DtRow) Resize(newWidth int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if newWidth < 0 {
		return uc.NewErrInvalidParameter(
			"newWidth",
			errors.New("value must be greater than 0"),
		)
	}

	if newWidth == r.width {
		return nil
	}

	if newWidth < r.width {
		r.cells = r.cells[:newWidth]
	} else {
		r.cells = append(r.cells, make([]*DtCell, newWidth-r.width)...)
	}

	r.width = newWidth

	return nil
}

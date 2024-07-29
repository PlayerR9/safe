package DtTable

import (
	"errors"
	"fmt"

	uc "github.com/PlayerR9/lib_units/common"
	rws "github.com/PlayerR9/safe/RWSafe"
)

// DtTable represents a table of cells.
// It is safe for concurrent use.
type DtTable struct {
	// height and width represent the height and width of the table, respectively.
	height, width *rws.Safe[int]

	// rows is a slice of rows in the table.
	rows []*DtRow
}

// GetCellAt returns the cell at the given coordinates.
// If the coordinates are out of bounds, it returns nil.
//
// Parameters:
//   - x: The x-coordinate.
//   - y: The y-coordinate.
//
// Returns:
//   - *DtCell: The cell at the given coordinates.
func (dt *DtTable) GetCellAt(x, y int) *DtCell {
	if y < 0 || y >= dt.height.Get() {
		return nil
	} else if x < 0 || x >= dt.width.Get() {
		return nil
	}

	return dt.rows[y].GetCellAt(x)
}

// GetWidth returns the width of the table.
//
// Returns:
//   - int: The width of the table.
func (dt *DtTable) GetWidth() int {
	return dt.width.Get()
}

// GetHeight returns the height of the table.
//
// Returns:
//   - int: The height of the table.
func (dt *DtTable) GetHeight() int {
	return dt.height.Get()
}

// SetCellAt sets the cell at the given coordinates.
//
// Parameters:
//   - x: The x-coordinate.
//   - y: The y-coordinate.
//   - cell: The cell to set.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if x and y are out of bounds.
func (dt *DtTable) SetCellAt(x, y int, cell *DtCell) error {
	height := dt.height.Get()
	width := dt.width.Get()

	if y < 0 || y >= height {
		return uc.NewErrInvalidParameter(
			"y",
			uc.NewErrOutOfBounds(y, 0, height),
		)
	} else if x < 0 || x >= width {
		return uc.NewErrInvalidParameter(
			"x",
			uc.NewErrOutOfBounds(x, 0, width),
		)
	}

	err := dt.rows[y].SetCell(cell, x)
	if err != nil {
		panic(fmt.Errorf("error setting cell: %w", err))
	}

	return nil
}

// NewDtTable creates a new table with the given height and width.
//
// Parameters:
//   - height: The height of the table.
//   - width: The width of the table.
//
// Returns:
//   - *DtTable: A pointer to the new table.
//   - error: An error of type *uc.ErrInvalidParameter if height or
//     width is less than 0.
func NewDtTable(height, width int) (*DtTable, error) {
	if height < 0 {
		return nil, uc.NewErrInvalidParameter(
			"height",
			errors.New("value must be non-negative"),
		)
	} else if width < 0 {
		return nil, uc.NewErrInvalidParameter(
			"width",
			errors.New("value must be non-negative"),
		)
	}

	rows := make([]*DtRow, height)

	for i := 0; i < height; i++ {
		row, err := NewDtRow(width)
		if err != nil {
			panic(fmt.Errorf("error creating row: %w", err))
		}

		rows[i] = row
	}

	return &DtTable{
		height: rws.NewSafe(height),
		width:  rws.NewSafe(width),
		rows:   rows,
	}, nil
}

// TransformIntoTable transforms a slice of cells into a table.
//
// Parameters:
//   - highlights: The slice of cells to transform.
//
// Returns:
//   - *DtTable: A pointer to the new table.
//   - error: An error of type *ErrInvalidCharacter if an invalid character is found.
func TransformIntoTable(highlights []DtCell) (*DtTable, error) {
	table := &DtTable{
		rows: make([]*DtRow, 0),
	}

	if len(highlights) == 0 {
		table.height = rws.NewSafe(0)
		table.width = rws.NewSafe(0)

		return table, nil
	}

	row, err := NewDtRow(0)
	if err != nil {
		panic(fmt.Errorf("error creating row: %w", err))
	}

	for _, hl := range highlights {
		switch hl.First {
		case '\n':
			table.rows = append(table.rows, row)

			row, err = NewDtRow(0)
			if err != nil {
				panic(fmt.Errorf("error creating row: %w", err))
			}
		// case '\t', '\r', '\b', '\f', '\v', '\a':
		// 	return nil, NewErrInvalidCharacter(hl.Content)
		case ' ':
			row.Append(nil)
		default:
			row.Append(&hl)
		}
	}

	if row.GetWidth() > 0 {
		table.rows = append(table.rows, row)
	}

	table.height = rws.NewSafe(len(table.rows))

	// Fix the sizes of the table.
	width := 0

	for _, row := range table.rows {
		rowWidth := row.GetWidth()

		if rowWidth > width {
			width = rowWidth
		}
	}

	table.width = rws.NewSafe(width)

	for i, row := range table.rows {
		if row.GetWidth() == width {
			continue
		}

		newRow, err := NewDtRow(width)
		if err != nil {
			panic(fmt.Errorf("error creating row %d: %w", i, err))
		}

		err = newRow.SetCells(row.cells, 0)
		if err != nil {
			panic(fmt.Errorf("error setting cells for row %d: %w", i, err))
		}

		table.rows[i] = newRow
	}

	return table, nil
}

// ResizeHeight resizes the height of the table.
//
// Parameters:
//   - newHeight: The new height of the table.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if newHeight is less than 0.
func (dt *DtTable) ResizeHeight(newHeight int) error {
	if newHeight < 0 {
		return uc.NewErrInvalidParameter(
			"newHeight",
			errors.New("value must be non-negative"),
		)
	}

	oldHeight := dt.height.Get()

	if newHeight == oldHeight {
		return nil
	}

	if newHeight < oldHeight {
		dt.rows = dt.rows[:newHeight]
	} else {
		width := dt.width.Get()

		for i := oldHeight; i < newHeight; i++ {
			row, err := NewDtRow(width)
			if err != nil {
				panic(fmt.Errorf("error creating row: %w", err))
			}

			dt.rows = append(dt.rows, row)
		}
	}

	dt.height.Set(newHeight)

	return nil
}

// ResizeWidth resizes the width of the table.
//
// Parameters:
//   - newWidth: The new width of the table.
//
// Returns:
//   - error: An error of type *uc.ErrInvalidParameter if newWidth is less than 0.
func (dt *DtTable) ResizeWidth(newWidth int) error {
	if newWidth < 0 {
		return uc.NewErrInvalidParameter(
			"newWidth",
			errors.New("value must be non-negative"),
		)
	}

	oldWidth := dt.width.Get()

	if newWidth == oldWidth {
		return nil
	}

	for _, row := range dt.rows {
		err := row.Resize(newWidth)
		if err != nil {
			panic(fmt.Errorf("error resizing row: %w", err))
		}
	}

	dt.width.Set(newWidth)

	return nil
}

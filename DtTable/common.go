package DtTable

type WriteOnlyDTer interface {
	// GetCellAt returns the cell at the given coordinates.
	// If the coordinates are out of bounds, it returns nil.
	//
	// Parameters:
	//   - x: The x-coordinate.
	//   - y: The y-coordinate.
	//
	// Returns:
	//   - *DtCell: The cell at the given coordinates.
	GetCellAt(x, y int) *DtCell

	// GetWidth returns the width of the table.
	//
	// Returns:
	//   - int: The width of the table.
	GetWidth() int

	// GetHeight returns the height of the table.
	//
	// Returns:
	//   - int: The height of the table.
	GetHeight() int

	// SetCellAt sets the cell at the given coordinates.
	//
	// Parameters:
	//   - x: The x-coordinate.
	//   - y: The y-coordinate.
	//   - cell: The cell to set.
	//
	// Returns:
	//   - error: An error of type *ers.ErrInvalidParameter if x and y are out of bounds.
	SetCellAt(x, y int, cell *DtCell) error

	// SignalChange signals a change to the table.
	SignalChange()
}

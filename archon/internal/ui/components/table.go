package components

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableComponent wraps bubbles.table with common styling and mouse support
type TableComponent struct {
	table        table.Model
	headerHeight int
	menuHeight   int
	tableStartY  int
	rowCount     int
}

// NewTableComponent creates a new table component
func NewTableComponent(columns []table.Column, rows []table.Row) *TableComponent {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10), // Default height, will be adjusted
	)

	// Apply default styles inline to avoid circular import
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Align(lipgloss.Left)
	s.Selected = lipgloss.NewStyle().
		Bold(true).
		Underline(true)
	s.Cell = lipgloss.NewStyle().
		Align(lipgloss.Left)
	t.SetStyles(s)

	return &TableComponent{
		table:        t,
		headerHeight: 1, // App header
		menuHeight:   1, // Navigation menu
		tableStartY:  4, // header + menu + table title + table header
		rowCount:     len(rows),
	}
}

// View returns the rendered table
func (c *TableComponent) View() string {
	return c.table.View()
}

// SetCursor sets the selected row index
func (c *TableComponent) SetCursor(index int) {
	if index >= 0 && index < c.rowCount {
		c.table.SetCursor(index)
	}
}

// GetCursor returns the current selected row index
func (c *TableComponent) GetCursor() int {
	return c.table.Cursor()
}

// HandleMouseClick calculates which row was clicked based on Y coordinate
// Returns the row index, or -1 if click was outside the table
func (c *TableComponent) HandleMouseClick(clickY int) int {
	// TODO: Implement click-to-row mapping
	// Steps:
	// 1. Check if clickY is within table bounds
	// 2. Calculate row index: rowIndex = clickY - c.tableStartY
	// 3. Validate row index is within 0 to c.rowCount-1
	// 4. Return row index or -1

	// Example:
	// if clickY < c.tableStartY {
	//     return -1 // Clicked above table
	// }
	//
	// rowIndex := clickY - c.tableStartY
	// if rowIndex >= 0 && rowIndex < c.rowCount {
	//     return rowIndex
	// }
	//
	// return -1

	return -1
}

// SetRows updates the table with new rows
func (c *TableComponent) SetRows(rows []table.Row) {
	c.table.SetRows(rows)
	c.rowCount = len(rows)
}

// SetHeight adjusts the table height based on available space
func (c *TableComponent) SetHeight(windowHeight int) {
	// Account for chrome (header, menu, status bar, margins)
	const chrome = 6
	height := windowHeight - chrome
	if height < 5 {
		height = 5
	}
	c.table.SetHeight(height)
}

// MoveUp moves the selection up by one row
func (c *TableComponent) MoveUp() {
	c.table.MoveUp(1)
}

// MoveDown moves the selection down by one row
func (c *TableComponent) MoveDown() {
	c.table.MoveDown(1)
}

// Update forwards TEA messages to the underlying table
func (c *TableComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	c.table, cmd = c.table.Update(msg)
	return cmd
}

// GetSelectedRow returns the currently selected row data
func (c *TableComponent) GetSelectedRow() table.Row {
	return c.table.SelectedRow()
}

// SetWidth adjusts the table width and proportionally scales columns
func (c *TableComponent) SetWidth(width int) {
	c.table.SetWidth(width)
}

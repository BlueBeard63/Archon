package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a single form field
type FormField struct {
	Label       string
	Placeholder string
	Input       textinput.Model
	Required    bool
	YPosition   int // Y coordinate for mouse click detection
}

// FormComponent manages a multi-field form with keyboard and mouse navigation
type FormComponent struct {
	Fields       []FormField
	FocusedIndex int
	startY       int // Y coordinate where form starts
	fieldSpacing int // Vertical spacing between fields
}

// NewFormComponent creates a new form with the given fields
func NewFormComponent(labels []string, placeholders []string, required []bool) *FormComponent {
	fields := make([]FormField, len(labels))

	for i := 0; i < len(labels); i++ {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.CharLimit = 256

		// Focus the first field by default
		if i == 0 {
			ti.Focus()
		}

		fields[i] = FormField{
			Label:       labels[i],
			Placeholder: placeholders[i],
			Input:       ti,
			Required:    required[i],
		}
	}

	return &FormComponent{
		Fields:       fields,
		FocusedIndex: 0,
		startY:       5, // After header, menu, title
		fieldSpacing: 3, // Label + input + blank line
	}
}

// View renders the form
func (f *FormComponent) View() string {
	var b strings.Builder

	// Define styles inline to avoid circular import
	labelStyle := lipgloss.NewStyle().Bold(true)

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)

	inputFocusedStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		Bold(true).
		Padding(0, 1)

	for i, field := range f.Fields {
		// Render label
		label := field.Label
		if field.Required {
			label += " *"
		}
		b.WriteString(labelStyle.Render(label) + "\n")

		// Render input field
		inputView := field.Input.View()
		if i == f.FocusedIndex {
			inputView = inputFocusedStyle.Render(inputView)
		} else {
			inputView = inputStyle.Render(inputView)
		}
		b.WriteString(inputView + "\n\n")

		// Update Y position for mouse click detection
		f.Fields[i].YPosition = f.startY + (i * f.fieldSpacing)
	}

	return b.String()
}

// Focus focuses a specific field by index
func (f *FormComponent) Focus(index int) {
	if index < 0 || index >= len(f.Fields) {
		return
	}

	// Blur current field
	f.Fields[f.FocusedIndex].Input.Blur()

	// Focus new field
	f.FocusedIndex = index
	f.Fields[f.FocusedIndex].Input.Focus()
}

// Next moves focus to the next field (Tab key)
func (f *FormComponent) Next() {
	next := (f.FocusedIndex + 1) % len(f.Fields)
	f.Focus(next)
}

// Previous moves focus to the previous field (Shift+Tab)
func (f *FormComponent) Previous() {
	prev := f.FocusedIndex - 1
	if prev < 0 {
		prev = len(f.Fields) - 1
	}
	f.Focus(prev)
}

// HandleMouseClick determines which field was clicked based on Y coordinate
// Returns the field index, or -1 if click was outside fields
func (f *FormComponent) HandleMouseClick(clickY int) int {
	// TODO: Implement click-to-field mapping
	// Steps:
	// 1. Iterate through fields
	// 2. Check if clickY matches field.YPosition (with tolerance)
	// 3. Return field index or -1

	// Example:
	// for i, field := range f.Fields {
	//     // Allow clicking on label or input (±1 row tolerance)
	//     if clickY >= field.YPosition && clickY <= field.YPosition+2 {
	//         return i
	//     }
	// }
	// return -1

	return -1
}

// Input handles character input for the focused field
func (f *FormComponent) Input(char rune) {
	if f.FocusedIndex >= 0 && f.FocusedIndex < len(f.Fields) {
		// TODO: Use textinput's Update method properly
		// f.Fields[f.FocusedIndex].Input.SetValue(
		//     f.Fields[f.FocusedIndex].Input.Value() + string(char),
		// )
	}
}

// Backspace removes the last character from the focused field
func (f *FormComponent) Backspace() {
	if f.FocusedIndex >= 0 && f.FocusedIndex < len(f.Fields) {
		// TODO: Use textinput's Update method properly
		// value := f.Fields[f.FocusedIndex].Input.Value()
		// if len(value) > 0 {
		//     f.Fields[f.FocusedIndex].Input.SetValue(value[:len(value)-1])
		// }
	}
}

// GetValues returns all field values as a slice
func (f *FormComponent) GetValues() []string {
	values := make([]string, len(f.Fields))
	for i, field := range f.Fields {
		values[i] = field.Input.Value()
	}
	return values
}

// Validate checks if all required fields are filled
func (f *FormComponent) Validate() []string {
	var errors []string

	for _, field := range f.Fields {
		if field.Required && strings.TrimSpace(field.Input.Value()) == "" {
			errors = append(errors, field.Label+" is required")
		}
	}

	return errors
}

// Reset clears all field values
func (f *FormComponent) Reset() {
	for i := range f.Fields {
		f.Fields[i].Input.SetValue("")
	}
	f.Focus(0)
}

package components

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

var (
	// Button styles
	buttonStyleNoBorder = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Margin(0, 1)

	buttonStyle = buttonStyleNoBorder.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("36"))

	buttonPrimaryStyle = buttonStyle.Copy().
				Background(lipgloss.Color("36")).
				Foreground(lipgloss.Color("15"))

	buttonSecondaryStyle = buttonStyle.Copy().
				BorderForeground(lipgloss.Color("240"))

	// Compact button style for icon-only buttons
	buttonCompactStyleNoBorder = lipgloss.NewStyle().
					Bold(true).
					Padding(0, 1).
					Margin(0)

	buttonCompactStyle = buttonCompactStyleNoBorder.
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240"))
)

// Button represents a clickable button
type Button struct {
	ID      string
	Label   string
	Primary bool
	Border  bool
	Icon    bool
}

// Render renders the button without zones
func (b *Button) Render() string {
	var style lipgloss.Style

	// Use compact style for icon-only buttons
	if b.Icon {
		if b.Border {
			style = buttonCompactStyle
		} else {
			style = buttonCompactStyleNoBorder
		}
	} else if b.Primary {
		style = buttonPrimaryStyle
	} else {
		style = buttonSecondaryStyle
	}

	return style.Render(b.Label)
}

// RenderWithZone renders the button with a clickable zone
func (b *Button) RenderWithZone(zm *zone.Manager) string {
	if zm == nil {
		return b.Render()
	}

	var style lipgloss.Style

	// Use compact style for icon-only buttons
	if b.Icon {
		if b.Border {
			style = buttonCompactStyle
		} else {
			style = buttonCompactStyleNoBorder
		}
	} else if b.Primary {
		style = buttonPrimaryStyle
	} else {
		style = buttonSecondaryStyle
	}

	rendered := style.Render(b.Label)
	return zm.Mark("button:"+b.ID, rendered)
}

// ButtonGroup renders multiple buttons horizontally
type ButtonGroup struct {
	Buttons []Button
}

// Render renders the button group without zones
func (bg *ButtonGroup) Render() string {
	var buttons []string
	for _, btn := range bg.Buttons {
		buttons = append(buttons, btn.Render())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
}

// RenderWithZones renders the button group with clickable zones
func (bg *ButtonGroup) RenderWithZones(zm *zone.Manager) string {
	if zm == nil {
		return bg.Render()
	}

	var buttons []string
	for _, btn := range bg.Buttons {
		buttons = append(buttons, btn.RenderWithZone(zm))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
}

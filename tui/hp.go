package main

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Controlling HP
func manageHP(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.Screen, m.Choice = 0, 0
			return m, nil
		case "h", "left":
			switch m.Step {
			case 0:
				m.Choice--
				if m.Choice < 0 {
					m.Choice = 0
				}
			case 1:
				m.Choice--
				if m.Choice < 3 {
					m.Choice = 3
				}
			}
		case "l", "right":
			switch m.Step {
			case 0:
				m.Choice++
				if m.Choice > 2 {
					m.Choice = 2
				}
			}
		case "enter":
			switch m.Step {
			case 0:
				switch m.Choice {
				case 0, 1:
					m.Step++
					return m, nil
				case 2:
					RestHP()
					m.Screen = 0
					m.Choice = 0
					return m, nil
				}
			case 1:
				enterAmount := m.Input.Value()
				m.Input.Reset()
				amountToChange, err := strconv.Atoi(enterAmount)
				if err != nil {
					help = true
					return m, nil
				}
				if m.Choice == 0 {
					amountToChange = -amountToChange
				}
				Damage(amountToChange, false)
				m.Step, m.Screen = 0, 0
				return m, nil
			}
		}
		if m.Step == 1 {
			m.Input.Focus()
		}
		if m.Input.Focused() {
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func adjustHP(m model) string {
	c := m.Choice
	s := m.Step

	tpl := "Select what you intend to happen to your hp:\n\n"
	tpl += "%s\n\n"
	if s == 1 {
		tpl += lipgloss.NewStyle().Align(lipgloss.Center).Render(
			"By how much:\n\n",
			m.Input.View(),
			"\n\n",
		)
	}
	if help {
		tpl += lipgloss.NewStyle().Foreground(lipgloss.Color("#f00")).Render(
			"Please enter a single positive number",
		)
	}
	tpl += subtleStyle.Render(
		"h/l, left, right: select damage or increase",
	) + dotStyle + subtleStyle.Render(
		"enter: choose",
	) + dotStyle + subtleStyle.Render(
		"q, esc, ctrl+c: quit",
	)

	choices := fmt.Sprintf(
		lipgloss.JoinHorizontal(lipgloss.Center,
			hpOptionStyle("Decrease HP", c == 0),
			hpOptionStyle("Increase HP", c == 1),
			hpOptionStyle("Reset HP", c == 2),
		),
	)
	return fmt.Sprintf(tpl, choices)
}

func hpOptionStyle(label string, checked bool) string {
	selected := lipgloss.NewStyle().
		Background(lipgloss.Color("#003434")).
		// BorderForeground(lipgloss.Color("#008080")).
		// BorderStyle(lipgloss.RoundedBorder()).
		Padding(1)
	unselected := lipgloss.NewStyle().
		Padding(1)

	if checked {
		return selected.Render(label)
	}
	return unselected.Render(label)
}

func areYouSure(label string, checked bool) string {
	selected := lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("212"))
	if checked {
		return selected.Render(label)
	}
	return fmt.Sprint(label)
}

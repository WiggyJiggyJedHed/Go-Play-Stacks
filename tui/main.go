package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	dotChar = " • "
)

var (
	subtleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dotStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	checkboxStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	mainStyle      = lipgloss.NewStyle().MarginLeft(2)
	createNewSheet bool
	currentHP      string
)

func checkbox(label string, checked bool) string {
	if checked {
		return checkboxStyle.Render("[x] " + label)
	}
	return fmt.Sprintf("[ ] %s", label)
}

func errBoiler(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

type model struct {
	Choice   int
	Chosen   bool
	Quitting bool
	Screen   int
	Input    textinput.Model
	Step     int
}

// General Functions
func main() {
	var initialModel model
	files, err := ConfirmChar(".", ".stacks")
	errBoiler(err)
	switch {
	case len(files) > 1:
		log.Fatalln("Error: Too many characters in this directory. Only use one character.")
	case len(files) == 0:
		initialModel = model{0, false, false, 7, textinput.New(), 0}
	case len(files) == 1:
		initialModel = model{0, false, false, 0, textinput.New(), 0}
	}

	// Assign stats
	char := &character{}
	ReadAndAssignStats(files[0], char)
	currentHP = string(char.HP)

	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}

		if m.Screen == 4 { // Display stats
			m.Screen = 0
			m.Choice = 0
			return m, nil
		}

		if m.Screen == 5 || m.Screen == 6 { // Roll Results
			m.Screen = 0
			m.Choice = 0
			return m, nil
		}
	}

	if m.Screen == 0 || m.Screen == 1 || m.Screen == 2 { // Main Menu and Rollers
		return updateChoices(msg, m)
	}
	if m.Screen == 3 {
		if m.Step == 2 {
			return manageHP(msg, m)
		}
		return manageHP(msg, m)
	}
	if m.Screen == 7 { // Character Creation
		switch m.Step {
		case 0:
			return enterName(msg, m)
		default:
			return diceSelectionScreen(msg, m)
		}
	}
	return m, tea.Quit
}

func (m model) View() string {
	var s string
	if m.Quitting {
		return "\n See you later!\n\n"
	}
	switch m.Screen {
	case 0:
		s = mainMenu(m)
	case 1, 2:
		s = rollMenu(m)
	case 3:
		s = adjustHP(m)
	case 4:
		s = displayChar(m)
	case 5, 6:
		s = resultView(m)
	case 7:
		if m.Step == 0 {
			s = nameInput(m)
		} else {
			s = diceSelector(m)
		}
	}
	return mainStyle.Render("\n" + s + "\n\n")
}

// Home screen with options
func mainMenu(m model) string {
	c := m.Choice

	tpl := "Select from the following:\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render(
		"j/k, up/down: select",
	) + dotStyle + subtleStyle.Render(
		"enter: choose",
	) + dotStyle + subtleStyle.Render(
		"q, esc: quit",
	)

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n",
		checkbox("Roll for attack", c == 0),
		checkbox("Roll for skill", c == 1),
		checkbox("Manage HP", c == 2),
		checkbox("Print your sheet", c == 3),
		checkbox("Exit", c == 4),
	)

	return fmt.Sprintf(tpl, choices)
}

// For Rolling stats
func rollMenu(m model) string {
	c := m.Choice

	tpl := "Select which stat to roll:\n\n"
	tpl += "%s\n\n"
	tpl += subtleStyle.Render(
		"j/k, up/down: select",
	) + dotStyle + subtleStyle.Render(
		"enter: choose",
	) + dotStyle + subtleStyle.Render(
		"q, esc: quit",
	)

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n",
		checkbox("Strength", c == 0),
		checkbox("Dexterity", c == 1),
		checkbox("Smarts", c == 2),
		checkbox("Constituion", c == 3),
		checkbox("Charisma", c == 4),
	)
	return fmt.Sprintf(tpl, choices)
}

func resultView(m model) string {
	return Roller(m.Screen, m.Choice)
}

func chosenView(string string) string {
	return string
}

func displayChar(m model) string {
	return printCharSheet()
}

func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			switch m.Screen {
			case 0:
				m.Quitting = true
				return m, tea.Quit
			case 1, 2:
				m.Screen = 0
				m.Choice = 0
				return m, nil
			}
		case "j", "down":
			m.Choice++
			if m.Choice > 4 {
				m.Choice = 4
			}
		case "k", "up":
			m.Choice--
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "enter":
			if m.Screen == 0 {
				m.Screen = m.Choice + 1
				m.Choice = 0
			} else if m.Screen == 1 || m.Screen == 2 {
				m.Screen = m.Screen + 4
				resultView(m)
				return m, nil
			}
			return m.Update(nil)
		}
	}
	return m, nil
}

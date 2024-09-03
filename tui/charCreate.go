package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	dice       []int
	sheetStats map[string]dice
)

var (
	d4  dice = []int{4}
	d6  dice = []int{4, 6}
	d8  dice = []int{4, 6, 8}
	d10 dice = []int{4, 6, 8, 10}
	d12 dice = []int{4, 6, 8, 10, 12}

	filename    string
	name        string
	beginTyping bool
	usedDice    = []string{}
	currentDie  string
	help        bool
)

var predefinedDice = map[string]dice{
	"d4":  d4,
	"d6":  d6,
	"d8":  d8,
	"d10": d10,
	"d12": d12,
}

var stats = sheetStats{
	"str": nil,
	"dex": nil,
	"smt": nil,
	"con": nil,
	"cha": nil,
}

type character struct {
	name           string
	HP             int
	maxHP          int
	level          int
	characterStats sheetStats
	modifier       int // modifiers that are applied to all rolls
}

// Edit File at Specific Line

func File2lines(filepath string) ([]string, error) {
	f, err := os.Open(filepath)
	errBoiler(err)
	defer f.Close()
	return LinesFromReader(f)
}

func LinesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, nil
}

func InsertStringToFile(path, str string, index int) error {
	lines, err := File2lines(path)
	errBoiler(err)
	fileContent := ""
	for i, line := range lines {
		if i == index {
			fileContent += str
		}
		fileContent += line
		fileContent += "\n"
	}

	return os.WriteFile(path, []byte(fileContent), 0644)
}

// Character Sheet Editing

func startingHP(name string) {
	fileName := name + ".stacks"
	hp := 5 + len(stats["con"])
	output := "\nHP = " + fmt.Sprint(hp) + "\nMaxHP = " + fmt.Sprint(hp) + "\n"
	err := InsertStringToFile(fileName, output, 3)
	errBoiler(err)
}

func increaseLevel(c *character, filename string) {
	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	errBoiler(err)
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	errBoiler(scanner.Err())

	for i, line := range content {
		if strings.HasPrefix(line, "Level = ") {
			c.level++
			content[i] = fmt.Sprintf("Level = %d", c.level)
		}
	}

	// Move the file pointer to the beginning of the file and truncate the file
	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, 0644)
	errBoiler(err)
	defer file.Close()

	// Write the updated content back to the file
	writer := bufio.NewWriter(file)
	for _, line := range content {
		_, err := writer.WriteString(line + "\n")
		errBoiler(err)
	}
	err = writer.Flush()
	errBoiler(err)
}

func setStats(input string) string {
	// Generate new character file
	file, err := os.Create(input + ".stacks")
	errBoiler(err)
	defer file.Close()

	// Write charcter name and starting level
	_, err = file.WriteString(fmt.Sprintf("## %s\n\nLevel = 1\n\n", name))
	errBoiler(err)

	// Write character's startingHP
	assignedCon := usedDice[4]
	hp := 5 + len(predefinedDice[assignedCon])
	output := "HP = " + fmt.Sprint(hp) + "\nMaxHP = " + fmt.Sprint(hp) + "\n\n"
	_, err = file.WriteString(output)
	errBoiler(err)

	file.WriteString(fmt.Sprintf("str = %s\n", usedDice[0]))
	file.WriteString(fmt.Sprintf("dex = %s\n", usedDice[1]))
	file.WriteString(fmt.Sprintf("smt = %s\n", usedDice[2]))
	file.WriteString(fmt.Sprintf("con = %s\n", usedDice[3]))
	file.WriteString(fmt.Sprintf("cha = %s\n", usedDice[4]))

	return file.Name()
}

type inputModel struct {
	input  textinput.Model
	step   int
	cursor int
}

// Step 1 where players enter their character's name
func NewForm(m *model) {
	nameIdeas := [...]string{
		"Randy Namo",
		"Hugh Mungus",
		"Shaquille O'Heal",
		"Conner the Poly Bomber",
		"Barak Ojama",
		"Dixie Normus",
		"Nami Dafuq",
		"Mister Fister",
	}
	name = nameIdeas[rand.Intn(len(nameIdeas))]

	m.Input = textinput.New()
	m.Input.Placeholder = name
	m.Input.Focus()
}

func enterName(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	if !beginTyping {
		NewForm(&m)
		beginTyping = true
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			switch m.Step {
			case 0:
				if len(m.Input.Value()) == 0 {
					m.Input.SetValue(name)
				}
				name = m.Input.Value()
				m.Input.Reset()
				m.Step++
				return m, nil
			}
		}
		if m.Input.Focused() {
			m.Input, cmd = m.Input.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func nameInput(m model) string {
	tpl := subtleStyle.Render(
		"enter: choose",
	) + dotStyle + subtleStyle.Render(
		"esc, ctrl+c: quit",
	)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"Enter the name of your character",
		m.Input.View(),
		tpl,
	)
}

// Step 2: Assign the dice to the character

func diceSelector(m model) string {
	c := m.Choice
	switch m.Step {
	case 1:
		currentDie = "Strength"
	case 2:
		currentDie = "Dexterity"
	case 3:
		currentDie = "Smarts"
	case 4:
		currentDie = "Constitution"
	case 5:
		currentDie = "Charisma"
	}
	selection := lipgloss.NewStyle().Bold(true).Padding(3, 0)

	tpl := "Select which die you want to assign to a stat:\n\n"
	tpl += selection.Render(currentDie) + "\n\n"
	tpl += "%s\n\n"
	if help {
		unavailable := lipgloss.NewStyle().Foreground(lipgloss.Color("#f00"))
		tpl += unavailable.Render("That die is unavailable. Select another") + "\n"
	}
	tpl += subtleStyle.Render(
		"h/l, left/right: select",
	) + dotStyle + subtleStyle.Render(
		"enter: choose",
	) + dotStyle + subtleStyle.Render(
		"q, esc: quit",
	)

	choices := fmt.Sprintf(
		"%s\t%s\t%s\t%s\t%s\t",
		displayAvailableDice("d4", c == 0),
		displayAvailableDice("d6", c == 1),
		displayAvailableDice("d8", c == 2),
		displayAvailableDice("d10", c == 3),
		displayAvailableDice("d12", c == 4),
	)
	return fmt.Sprintf(tpl, choices)
}

func diceSelectionScreen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "h", "left":
			m.Choice--
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "l", "right":
			m.Choice++
			if m.Choice > 4 {
				m.Choice = 4
			}
		case "enter":
			help = false
			if !isDieValid(m) {
				help = true
				return m, nil
			}
			acceptDiceChange(m)
			if m.Step != 5 {
				m.Step++
				return m, nil
			}
			setStats(name)
			m.Step, m.Screen, m.Choice = 0, 0, 0
			return m, nil
		}
	}
	return m, nil
}

func acceptDiceChange(m model) {
	switch m.Choice {
	case 0:
		usedDice = append(usedDice, "d4")
	case 1:
		usedDice = append(usedDice, "d6")
	case 2:
		usedDice = append(usedDice, "d8")
	case 3:
		usedDice = append(usedDice, "d10")
	case 4:
		usedDice = append(usedDice, "d12")
	}
}

func displayAvailableDice(label string, checked bool) string {
	unavailable := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")) // gray
	selected := lipgloss.NewStyle().Underline(true).Foreground(lipgloss.Color("212"))
	both := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Underline(true)

	if slices.Contains(usedDice, label) && checked {
		return both.Render(label)
	} else if slices.Contains(usedDice, label) && !checked {
		return unavailable.Render(label)
	}

	if checked {
		return selected.Render(label)
	}
	return fmt.Sprint(label)
}

func isDieValid(m model) bool {
	var s string
	switch m.Choice {
	case 0:
		s = "d4"
	case 1:
		s = "d6"
	case 2:
		s = "d8"
	case 3:
		s = "d10"
	case 4:
		s = "d12"
	}

	if slices.Contains(usedDice, s) {
		return false
	}
	return true
}

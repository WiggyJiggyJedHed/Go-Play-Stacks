package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	barStyle    = strings.Repeat("=", 24)
	starStyle   = strings.Repeat("*", 24)
	borderColor = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	crit        = lipgloss.NewStyle().Foreground(lipgloss.Color("31")).SetString("Critical Success")
)

// Core Roller function
func Roller(method int /*attack or skill*/, stat int /*which stat is rolled*/) string {
	files, err := ConfirmChar(".", ".stacks")
	errBoiler(err)
	char := &character{}
	ReadAndAssignStats(files[0], char)

	dice := assignRolls(char, stat)
	x := method - 5
	if x == 0 {
		return attack(dice)
	} else if x == 1 {
		return skill(dice)
	}
	return "Error in read"
}

// Reset HP
func RestHP() error {
	files, err := ConfirmChar(".", ".stacks")
	errBoiler(err)
	c := &character{}
	ReadAndAssignStats(files[0], c)
	file, err := os.OpenFile(files[0], os.O_RDWR, 0644)
	errBoiler(err)
	defer file.Close()

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	errBoiler(scanner.Err())

	for i, line := range content {
		if strings.HasPrefix(line, "HP = ") {
			content[i] = fmt.Sprintf("HP = %d", c.maxHP)
			c.HP = c.maxHP
		}
	}

	// Move the file pointer to the beginning of the file and truncate the file
	file, err = os.OpenFile(files[0], os.O_WRONLY|os.O_TRUNC, 0644)
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
	return err
}

func printCharSheet() string {
	files, err := ConfirmChar(".", ".stacks")
	errBoiler(err)
	char := &character{}
	ReadAndAssignStats(files[0], char)

	file, err := os.Open(files[0])
	errBoiler(err)
	defer file.Close()

	output, err := io.ReadAll(file)
	return fmt.Sprint(barStyle + "\n" + string(output) + barStyle)
}

// Change the HP without resetting it
func Damage(dmg int, reset bool) string {
	files, err := ConfirmChar(".", ".stacks")
	errBoiler(err)
	c := &character{}
	file, err := os.OpenFile(files[0], os.O_RDWR, 0644)
	errBoiler(err)
	defer file.Close()

	if reset {
		dmg = c.maxHP
	}

	var content []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}
	errBoiler(scanner.Err())

	for i, line := range content {
		if strings.HasPrefix(line, "HP = ") {
			content[i] = fmt.Sprintf("HP = %d", c.HP+dmg)
			c.HP += dmg
		}
	}

	// Move the file pointer to the beginning of the file and truncate the file
	file, err = os.OpenFile(files[0], os.O_WRONLY|os.O_TRUNC, 0644)
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
	return string(c.HP)
}

// Read and Assign stats
func ConfirmChar(root, ext string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		errBoiler(err)
		if !info.IsDir() && strings.HasSuffix(info.Name(), ext) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func ReadAndAssignStats(filename string, char *character) {
	file, err := os.Open(filename)
	errBoiler(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	char.characterStats = make(sheetStats)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Parse the character's name
		if strings.HasPrefix(line, "## ") {
			char.name = strings.TrimSpace(line[3:])
			continue
		}

		if strings.Contains(line, " = ") {
			parts := strings.Split(line, " = ")
			if len(parts) != 2 {
				fmt.Println("Invalid line format:", line)
				continue
			}
			stat, die := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

			switch stat {
			case "Level":
				char.level, err = strconv.Atoi(die)
				errBoiler(err)
			case "HP":
				char.HP, err = strconv.Atoi(die)
				errBoiler(err)
			case "MaxHP":
				char.maxHP, err = strconv.Atoi(die)
				errBoiler(err)
			default:
				if selectedDice, ok := predefinedDice[die]; ok {
					char.characterStats[stat] = selectedDice
				}
			}
		}
	}
}

func assignRolls(c *character, index int) dice {
	var x dice
	switch index {
	case 0:
		x = c.characterStats["str"]
	case 1:
		x = c.characterStats["dex"]
	case 2:
		x = c.characterStats["smt"]
	case 3:
		x = c.characterStats["con"]
	case 4:
		x = c.characterStats["cha"]
	}
	return x
}

// Roll functions
func roll(n int) int {
	return rand.Intn(n) + 1
}

func reRoll(n int) int {
	roll := roll(n)
	if roll == n {
		return n + reRoll(n)
	}
	return roll
}

func skill(dice dice) string {
	output := borderColor.Render(fmt.Sprintf("\n%v", barStyle)) + "\n"
	for _, d := range dice {
		result := reRoll(d)
		if result > d {
			output += fmt.Sprintf("d%v: %v %v\n", d, result+2, crit)
			continue
		}
		output += fmt.Sprintf("d%v: %v\n", d, result+2)
	}
	output += borderColor.Render(fmt.Sprintf("%v\n\n", barStyle))
	return output
}

func attack(dice dice) string {
	output := borderColor.Render(fmt.Sprintf("\n%v", barStyle)) + "\n"
	for _, d := range dice {
		result := roll(d)
		if result == d {
			output += fmt.Sprintf("d%v: %v %v\n", d, result+2, crit)
			continue
		}
		output += fmt.Sprintf("d%v: %v\n", d, result+2)
	}
	output += borderColor.Render(fmt.Sprintf("%v\n\n", barStyle))
	return output
}

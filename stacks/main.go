package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"stacks/charCreate"
	"strconv"
	"strings"
)

const (
	selectAction = `Select from the following:
  1. Roll for attack
  2. Roll for skill
  3. Increase/Decrease HP
  4. Rest (Reset HP)
  5. Print your stats
  6. Exit`

	selectStat = `  1. Strength
  2. Dexterity
  3. Smarts
  4. Constitution
  5. Charisma`
)

type (
	dice       []int
	inventory  []item
	sheetStats map[string]dice
)

var (
	d4  dice = []int{4}
	d6  dice = []int{4, 6}
	d8  dice = []int{4, 6, 8}
	d10 dice = []int{4, 6, 8, 10}
	d12 dice = []int{4, 6, 8, 10, 12}

	barStyle  = strings.Repeat("=", 24)
	starStyle = strings.Repeat("*", 24)
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
	spells         []spell
	inventory      inventory
	equipped       []item
	modifier       int // modifiers that are applied to all rolls
}

type spell struct{}

type item struct {
	weapon      bool
	description string
	modifier    int
	stat        string
}

func errBoiler(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Read and Assign stats

func readAndAssignStats(filename string, char *character) {
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

func skill(dice dice) {
	fmt.Printf("\n%v\n", strings.Repeat("=", 24))
	for _, d := range dice {
		result := reRoll(d)
		if result > d {
			fmt.Printf("d%v: %v  Critical Success\n", d, result)
			continue
		}
		fmt.Printf("d%v: %v\n", d, result)
	}
	fmt.Printf("%v\n\n", barStyle)
}

func attack(dice dice) {
	fmt.Printf("\n%v\n", strings.Repeat("=", 24))
	for _, d := range dice {
		result := roll(d)
		if result == d {
			fmt.Printf("d%v: %v  Critical Success\n", d, result)
			continue
		}
		fmt.Printf("d%v: %v\n", d, result)
	}
	fmt.Printf("%v\n\n", barStyle)
}

func damage(c *character, dmg int, filename string) {
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
		if strings.HasPrefix(line, "HP = ") {
			content[i] = fmt.Sprintf("HP = %d", c.HP+dmg)
			c.HP += dmg
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

func resetHP(c *character, filename string) {
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
		if strings.HasPrefix(line, "HP = ") {
			content[i] = fmt.Sprintf("HP = %d", c.maxHP)
			c.HP = c.maxHP
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

// MAIN

func main() {
	files, err := charCreate.ConfirmChar(".", ".stacks")
	errBoiler(err)
	switch {
	case len(files) > 1:
		log.Fatalln("Error: Too many characters in this directory. Only use one character.")
		return
	case len(files) == 0:
		path := charCreate.CharCreate()
		files = append(files, path)
	}

	// Assign stats
	char := &character{}
	readAndAssignStats(files[0], char)

	// User select action
	for {
		fmt.Println(selectAction)
		var selection string
		var input string
		_, err = fmt.Scanln(&input)
		errBoiler(err)

		switch input {
		case "1": // Attack
			fmt.Println(selectStat)
			_, err = fmt.Scanln(&selection)
			switch selection {
			case "1": // Strength
				attack(char.characterStats["str"])
			case "2": // Dexterity
				attack(char.characterStats["dex"])
			case "3": // Smarts
				attack(char.characterStats["smt"])
			case "4": // Constitution
				attack(char.characterStats["con"])
			case "5": // Charisma
				attack(char.characterStats["cha"])
			}

		case "2": // Skill
			fmt.Println(selectStat)
			_, err = fmt.Scanln(&selection)
			switch selection {
			case "1": // Strength
				skill(char.characterStats["str"])
			case "2": // Dexterity
				skill(char.characterStats["dex"])
			case "3": // Smarts
				skill(char.characterStats["smt"])
			case "4": // Constitution
				skill(char.characterStats["con"])
			case "5": // Charisma
				skill(char.characterStats["cha"])
			}

		case "3": // Take damage / regain life
			fmt.Println("What are you gaining/losing? (include a negative '-' if losing)")
			_, err = fmt.Scanln(&selection)
			dmg, err := strconv.Atoi(selection)
			errBoiler(err)
			damage(char, dmg, files[0])
			fmt.Printf("\nYour HP is now at %d\n", char.HP)

		case "4": // Reset HP
			fmt.Printf("Are you sure? Current HP = %d\n", char.HP)
			_, err = fmt.Scanln(&selection)
			if selection == "y" || selection == "yes" {
				resetHP(char, files[0])
				fmt.Printf(
					"\n%s\n* HP reset to max (%d) *\n%s\n",
					starStyle,
					char.HP,
					starStyle,
				)
			} else {
				continue
			}

		case "5":
			file, err := os.Open(files[0])
			errBoiler(err)
			defer file.Close()

			output, err := io.ReadAll(file)
			fmt.Print("\n" + barStyle + "\n" + string(output) + barStyle + "\n\n")

		case "6":
			fmt.Println(char.characterStats["smt"])
			os.Exit(30)
		}
	}
}

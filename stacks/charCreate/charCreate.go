package charCreate

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// Err Boiler

func errBoiler(err error) {
	if err != nil {
		log.Fatalln(err)
	}
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

// Steps in creation
func enterName() string {
	fmt.Println("Enter your character's name:")
	var name string
	_, err := fmt.Scanln(&name)
	errBoiler(err)
	fileName := name + ".stacks"
	file, err := os.Create(fileName)
	errBoiler(err)
	defer file.Close()
	return name
}

func setStats(name string) string {
	fileName := name + ".stacks"
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
	errBoiler(err)
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("## %s\n\nLevel = 1\n\n", name))
	errBoiler(err)

	fmt.Println("Assign the dice to each of your stats.")
	for stat := range stats {
		fmt.Printf("Enter die for %s (d4, d6, d8, d10, d12): ", stat)
		var input string
		_, err := fmt.Scanln(&input)
		errBoiler(err)

		if selectedDice, ok := predefinedDice[input]; ok {
			stats[stat] = selectedDice
			lineEntry := fmt.Sprintf("%s = %v\n", stat, input)
			_, err = file.WriteString(lineEntry)
			errBoiler(err)
		} else {
			log.Println("Invalid input. You'll have to go back to your sheet to fix that. Remember, enter valid die.")
			continue
		}
	}

	return file.Name()
}

func CharCreate() string {
	name := enterName()
	filepath := setStats(name)
	startingHP(name)
	return filepath
}

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

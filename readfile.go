package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// https://freshman.tech/snippets/go/roman-numerals/ Task 6
func readTextFilesFromFolder(filePath string) (string, error) {
	// Read the file content
	// content, err := ioutil.ReadFile(filePath)
	contentBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	fmt.Println("Raw bytes:", contentBytes)
	return string(contentBytes), nil
}

// https://freshman.tech/snippets/go/roman-numerals/ task 6

func romanToInteger(roman string) (int, error) {
	romanMap := map[string]int{
		"M":  1000,
		"CM": 900,
		"D":  500,
		"CD": 400,
		"C":  100,
		"XC": 90,
		"L":  50,
		"XL": 40,
		"X":  10,
		"IX": 9,
		"V":  5,
		"IV": 4,
		"I":  1,
	}

	number := 0
	roman = strings.ToUpper(roman)
	for i := 0; i < len(roman); i++ {
		currentChar := string(roman[i])
		if currentChar == "\x00" {
			continue // Skip null character
		}
		currentValue, ok := romanMap[currentChar]
		if !ok {
			return 0, fmt.Errorf("invalid Roman numeral character: %q", currentValue)
		}
	}
	return number, nil
}

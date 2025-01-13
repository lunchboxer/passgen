package main

import (
	"bufio"
	"crypto/rand"
	"embed"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/atotto/clipboard"
)

//go:embed words.txt short-words.txt
var wordFiles embed.FS

func main() {
	// Parse command-line arguments
	numWords := flag.Int("w", 4, "Number of words in the password")
	separator := flag.String("s", "-", "Separator between words")
	length := flag.Int("l", 0, "Total length of the password (optional)")
	capitalize := flag.Bool("c", false, "Capitalize the first letter of each word")
	addNumber := flag.Bool("n", false, "Add a random number at the end")
	addSymbol := flag.Bool("y", false, "Add a non-word symbol at the end")
	clipboardFlag := flag.Bool("b", false, "Copy the password to the clipboard")
	help := flag.Bool("h", false, "Show this help message")
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	// Load words from embedded files
	words, err := loadWords("words.txt")
	if err != nil {
		fmt.Println("Error loading words:", err)
		return
	}
	shortWords, err := loadWords("short-words.txt")
	if err != nil {
		fmt.Println("Error loading short words:", err)
		return
	}

	// Generate the password
	password, err := generatePassword(words, shortWords, *numWords, *separator, *length, *capitalize, *addNumber, *addSymbol)
	if err != nil {
		fmt.Println("Error generating password:", err)
		return
	}

	// Output the password
	fmt.Println(password)

	// Copy the password to the clipboard if requested
	if *clipboardFlag {
		if err := clipboard.WriteAll(password); err != nil {
			fmt.Println("Error copying to clipboard:", err)
		} else {
			fmt.Println("Password copied to clipboard!")
		}
	}
}

// loadWords reads words from an embedded file and returns them as a slice
func loadWords(filename string) ([]string, error) {
	file, err := wordFiles.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	if len(words) == 0 {
		return nil, errors.New("no words found in the file")
	}

	return words, nil
}

// generatePassword generates a secure XKCD-style password
func generatePassword(words, shortWords []string, numWords int, separator string, length int, capitalize, addNumber, addSymbol bool) (string, error) {
	if numWords <= 0 {
		return "", errors.New("number of words must be greater than 0")
	}

	// Validate the length argument
	if length > 0 {
		minLength := numWords*3 + (numWords-1)*len(separator)  // Minimum length with 3-letter words
		maxLength := numWords*14 + (numWords-1)*len(separator) // Maximum length with 14-letter words
		if length < minLength {
			return "", fmt.Errorf("requested length %d is too short for %d words (minimum length is %d)", length, numWords, minLength)
		}
		if length > maxLength {
			return "", fmt.Errorf("requested length %d is too long for %d words (maximum length is %d)", length, numWords, maxLength)
		}
	}

	// If length is specified, find words that fit the length requirement
	var passwordWords []string
	if length > 0 {
		var err error
		passwordWords, err = findWordsForLength(words, shortWords, numWords, len(separator), length)
		if err != nil {
			return "", err
		}
	} else {
		// Otherwise, select random words
		for i := 0; i < numWords; i++ {
			index, err := getSecureRandomIndex(len(words))
			if err != nil {
				return "", err
			}
			passwordWords = append(passwordWords, words[index])
		}
	}

	// Capitalize words if requested
	if capitalize {
		for i, word := range passwordWords {
			passwordWords[i] = capitalizeFirstLetter(word)
		}
	}

	// Join words with the separator
	password := strings.Join(passwordWords, separator)

	// Add a random number at the end
	if addNumber {
		num, err := getSecureRandomNumber(0, 9)
		if err != nil {
			return "", err
		}
		password += fmt.Sprintf("%d", num)
	}

	// Add a non-word symbol at the end
	if addSymbol {
		symbols := []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '+', '=', '~', '`', '|', '\\', '/', '?', '>', '<', '.', ','}
		index, err := getSecureRandomIndex(len(symbols))
		if err != nil {
			return "", err
		}
		password += string(symbols[index])
	}

	return password, nil
}

// findWordsForLength selects words that fit the specified length
func findWordsForLength(words, shortWords []string, numWords, separatorLength, targetLength int) ([]string, error) {
	startTime := time.Now() // Record the start time
	var passwordWords []string

	for time.Since(startTime) < time.Second { // Try for up to 1 second
		passwordWords = nil // Reset the word list
		remainingLength := targetLength

		for i := 0; i < numWords; i++ {
			// Calculate the maximum word length allowed
			maxWordLength := remainingLength - (numWords-i-1)*separatorLength
			if maxWordLength <= 0 {
				break // This combination won't work; start over
			}

			// Use short words if the remaining length is small
			var wordList []string
			if maxWordLength <= 2 {
				wordList = shortWords
			} else {
				wordList = words
			}

			// Find a word that fits
			var word string
			for attempts := 0; attempts < 100; attempts++ { // Try up to 100 times for each word
				index, err := getSecureRandomIndex(len(wordList))
				if err != nil {
					return nil, err
				}
				candidate := wordList[index]
				if len(candidate) <= maxWordLength {
					word = candidate
					break
				}
			}

			if word == "" {
				break // Couldn't find a word; start over
			}

			passwordWords = append(passwordWords, word)
			remainingLength -= len(word)
			if i < numWords-1 {
				remainingLength -= separatorLength // Subtract the separator length for all but the last word
			}
		}

		// If we found a valid combination, return it
		if len(passwordWords) == numWords && remainingLength == 0 {
			return passwordWords, nil
		}
	}

	// If we couldn't find a valid combination within 1 second, return an error
	return nil, fmt.Errorf("could not find a valid combination of words for length %d", targetLength)
}

// getSecureRandomIndex generates a secure random index within a range
func getSecureRandomIndex(max int) (int, error) {
	if max <= 0 {
		return 0, errors.New("max must be greater than 0")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

// getSecureRandomNumber generates a secure random number within a range
func getSecureRandomNumber(min, max int) (int, error) {
	if min > max {
		return 0, errors.New("min must be less than or equal to max")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}

// capitalizeFirstLetter capitalizes the first letter of a string
func capitalizeFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(unicode.ToUpper(rune(s[0]))) + s[1:]
}

// printHelp displays the help message
const Version = "0.1.5"

func printHelp() {
	fmt.Printf(`
Password Generator v%s
Usage: passgen [options]

Options:
  -w, --words      Number of words in the password (default: 4)
  -s, --separator  Separator between words (default: '-')
  -l, --length     Total length of the password (optional)
  -c, --capitalize Capitalize the first letter of each word
  -n, --number     Add a random number at the end
  -y, --symbol     Add a non-word symbol at the end
  -b, --clipboard  Copy the password to the clipboard
  -v, --version    Show version information
  -h, --help       Show this help message
`, Version)
}

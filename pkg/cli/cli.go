package cli

import "fmt"

// PromptUser gets user input from command line
func PromptUser(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// PromptPassword gets password input from command line
func PromptPassword(prompt string) string {
	fmt.Print(prompt)
	var password string
	fmt.Scanln(&password)
	return password
}

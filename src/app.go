package main

import "fmt"

// App represents the main application and holds its dependencies
type App struct {
	config    *Config
	jira      *JiraClient
	generator *Generator
}

// NewApp creates a new instance of the application with its dependencies
func NewApp(config *Config) *App {
	return &App{
		config:    config,
		jira:      NewJiraClient(config),
		generator: NewGenerator(config),
	}
}

// Run executes the main application logic
func (app *App) Run(customPrompt string) error {
	prompt, err := app.getPrompt(customPrompt)
	if err != nil {
		return err
	}

	output, err := app.generator.Generate(prompt)
	if err != nil {
		return err
	}

	if err := handleOutput(output); err != nil {
		return err
	}

	return nil
}

// getPrompt determines the final prompt to use based on Jira availability and custom input
func (a *App) getPrompt(customPrompt string) (string, error) {
	// If Jira is not configured, use custom prompt
	if !a.jira.IsConfigured() {
		if customPrompt == "" {
			return "", fmt.Errorf("prompt is required when not using Jira")
		}
		return customPrompt, nil
	}

	// Try to get Jira ticket
	ticket, err := a.jira.SelectTicket()
	if err != nil {
		if customPrompt == "" {
			return "", fmt.Errorf("no prompt provided and Jira failed: %v", err)
		}
		fmt.Printf("Warning: Jira integration failed, falling back to custom prompt: %v\n", err)
		return customPrompt, nil
	}

	prompt := fmt.Sprintf("Jira ticket: %s\nSummary: %s\nDescription: %s\n",
		ticket.Key,
		ticket.Fields.Summary,
		ticket.Fields.Description)

	if customPrompt != "" {
		prompt += "\n" + customPrompt
	}

	return prompt, nil
}

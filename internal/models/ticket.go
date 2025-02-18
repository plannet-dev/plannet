package models

// Ticket represents a generic work item
type Ticket struct {
	ID          string
	Title       string
	Description string
	Status      string
	URL         string
}

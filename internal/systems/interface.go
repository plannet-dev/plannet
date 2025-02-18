package systems

import "plannet/internal/models"

// TicketSystem defines the interface for different ticket management systems
type TicketSystem interface {
	// Authenticate logs in and returns a token or session
	Authenticate(username, password string) error

	// GetAssignedTickets retrieves tickets assigned to the current user
	GetAssignedTickets() ([]models.Ticket, error)

	// Logout terminates the current session
	Logout() error
}

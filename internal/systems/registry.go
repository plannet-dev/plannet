package systems

// SystemInfo contains metadata about a ticket system
type SystemInfo struct {
	ID          string
	Name        string
	Description string
	RequiresURL bool
	URLExample  string
}

// Registry manages available ticket systems
type Registry struct {
	systems []SystemInfo
}

// NewRegistry creates a new system registry with default systems
func NewRegistry() *Registry {
	return &Registry{
		systems: []SystemInfo{
			{
				ID:          "jira",
				Name:        "Jira",
				Description: "Atlassian Jira - Project Management Tool",
				RequiresURL: true,
				URLExample:  "https://your-domain.atlassian.net",
			},
			{
				ID:          "asana",
				Name:        "Asana",
				Description: "Asana Project Management",
				RequiresURL: false,
				URLExample:  "",
			},
		},
	}
}

// GetSystems returns all available systems
func (r *Registry) GetSystems() []SystemInfo {
	return r.systems
}

// GetSystemInfo returns the system info for a given system ID
func (r *Registry) GetSystemInfo(id string) (*SystemInfo, bool) {
	for _, sys := range r.systems {
		if sys.ID == id {
			return &sys, true
		}
	}
	return nil, false
}

// Add registers a new system
func (r *Registry) Add(system SystemInfo) {
	r.systems = append(r.systems, system)
}

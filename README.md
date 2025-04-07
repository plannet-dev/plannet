# Plannet - LLM-Powered Planning Assistant

Plannet is a command-line tool that helps you stay on top of your workload and backlog from where you work, the command line.

You can bring your own LLM, or use ours at https://www.plannet.dev/

## Features

- 🤖 LLM integration for content generation
- 🎫 Optional Jira integration
- 🎨 Rich terminal UI for ticket selection
- ⚙️ Configurable to bring your own model
- 📋 Clipboard integration with customizable behavior

## Installation

### Prerequisites

- Go 1.19 or later
- A Jira account (optional)
- Access to an LLM API endpoint

### Building from Source

```bash
# Clone the repository
git clone https://github.com/plannet-ai/plannet.git
cd plannet

# Install dependencies
go mod download

# Build the binary
./build.sh
```

## Configuration

Create a `.plannetrc` file in your home directory with the following structure:

```json
{
  "base_url": "http://localhost:1234/v1/completions",
  "model": "your-model-name",
  "system_prompt": "Optional system prompt to guide the LLM",
  "headers": {
    "Authorization": "Bearer your-token"
  },
  "jira_url": "https://your-instance.atlassian.net",
  "jira_token": "your-jira-token",
  "jira_user": "your-email@company.com",
  "copy_preference": "ask-every-time"
}
```

### Configuration Fields

- **Required Fields:**

  - `base_url`: The LLM API endpoint
  - `model`: The model identifier to use
  - `headers`: API authentication headers

- **Optional Fields:**
  - `system_prompt`: A prompt that guides the LLM's behavior
  - `jira_url`: Your Jira instance URL
  - `jira_token`: Your Jira API token
  - `jira_user`: Your Jira username/email
  - `copy_preference`: How to handle clipboard copying (options: ask-every-time, ask-once, copy-automatically, do-not-copy)

## Usage

### Basic Usage

```bash
# Generate content with a custom prompt
plannet --prompt "Generate a test plan for a login feature"

# Use the generate command
plannet generate "Generate a test plan for a login feature"

# Use Jira integration (if configured)
plannet jira list

# Show version information
plannet --version

# Enable debug mode
plannet --debug
```

### Using with Jira

When Jira integration is configured, Plannet will:

1. Fetch your assigned tickets
2. Present an interactive selection interface
3. Use the selected ticket's information to generate content
4. Optionally append your custom prompt

### Output Management

- Generated content is displayed in the terminal
- Option to copy output to clipboard
- Color-coded output for better readability
- Configurable clipboard behavior

## Development

### Project Structure

```
.
├── main.go          # Application entry point
├── cmd/             # Command implementations using Cobra
│   ├── root.go      # Root command definition
│   ├── jira.go      # Jira-related commands
│   ├── track.go     # Work tracking commands
│   ├── generate.go  # Content generation command
│   └── ...          # Other command files
├── config/          # Configuration management
│   ├── config.go    # Configuration struct and functions
│   ├── copy_preference.go # Clipboard preference handling
│   └── config_test.go # Configuration tests
├── llm/             # LLM interaction
│   └── generator.go # LLM request handling
├── output/          # Output management
│   └── output.go    # Output display and clipboard handling
├── build/           # Build output directory
├── build.sh         # Build script
└── test_track.sh    # Test script for tracking feature
```

> **Note:** The `src/` directory contains an old implementation that is no longer used. See `src/README.md` for details.

### Adding New Features

1. Create a new command file in the `cmd/` directory
2. Add your command to the root command in `cmd/root.go`
3. Update the README.md with new features

### Running Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [Cobra](https://github.com/spf13/cobra) for CLI commands
- Uses [promptui](https://github.com/manifoldco/promptui) for interactive prompts
- Uses [fatih/color](https://github.com/fatih/color) for terminal colors

# Plannet

Plannet is a command-line tool for managing your daily tasks and projects. It integrates with various services like Jira and LLMs to help you stay organized and productive.

## Features

- Task management with ticket prefixes
- Jira integration for viewing and creating tickets
- LLM integration for getting help with tasks
- Git integration for managing branches
- Secure storage for API tokens
- Rate limiting for API calls
- Input validation and file path sanitization

## Installation

### Prerequisites

- Go 1.21 or later
- Git (optional, for git integration)

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/plannet-ai/plannet.git
cd plannet
```

2. Build the project:

```bash
go build
```

3. Add plannet to your PATH:

```bash
# For bash/zsh (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(pwd)

# For fish (add to ~/.config/fish/config.fish)
set -gx PATH $PATH (pwd)
```

## Usage

### Initial Setup

Run the initialization command to set up your configuration:

```bash
plannet init
```

This will guide you through setting up:

- Ticket prefixes for your projects
- Editor preferences
- Git integration settings
- Jira integration
- LLM integration

### Managing Tasks

Create a new task:

```bash
plannet task "Implement user authentication"
```

List your tasks:

```bash
plannet list
```

### Jira Integration

View your Jira tickets:

```bash
plannet jira list
```

View a specific ticket:

```bash
plannet jira view PROJ-123
```

Create a new ticket:

```bash
plannet jira create
```

### LLM Integration

Start an interactive session with the LLM:

```bash
plannet llm
```

Send a single prompt:

```bash
plannet llm --prompt "How do I implement rate limiting in Go?"
```

## Configuration

The configuration file is stored at `~/.plannetrc`. It contains your preferences and settings for various integrations.

## Security

Plannet implements several security features:

- Secure storage for API tokens using AES-GCM encryption
- Rate limiting for API calls to prevent abuse
- Input validation to prevent injection attacks
- File path sanitization to prevent path traversal attacks

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

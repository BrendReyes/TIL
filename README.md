# TIL (Today I Learned)

A CLI tool for capturing and reviewing things you learn, designed to help you retain information through spaced repetition.

## Features

- **Capture:** Add new insights instantly from your terminal.
- **Review:** Spaced Repetition System (SRS) to ensure long-term retention.
- **TUI:** Interactive Terminal User Interface for easy management.
- **Stats:** Track your learning progress over time.

## Installation

To install `til`, you must have [Go](https://go.dev/doc/install) installed on your system.

```bash
go install github.com/brendreyes/til@latest
```

### Path Configuration

Make sure your Go bin directory is in your system's `PATH` to run the `til` command from anywhere.

- **Linux/macOS:** Add `export PATH=$PATH:$(go env GOPATH)/bin` to your `.bashrc` or `.zshrc`.
- **Windows:** Add `%USERPROFILE%\go\bin` to your Environment Variables.

## Usage

```bash
til         # Show help and commands
til tui     # Start the interactive TUI
til add     # Add a new entry
til review  # Start a review session
til list    # List all entries
til stats   # Show learning statistics
```

## License

[MIT](LICENSE)

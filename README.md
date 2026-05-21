# TIL (Today I Learned)

This is just a simple cli/tui tool to store what you have learned, mainly for when studying, debugging, learned something new, etc. And it has a simple spaced repetition review just a way to not completely forget everything.

<img width="600" height="500" alt="demo" src="https://github.com/user-attachments/assets/9e6a9598-7f7b-4b61-a98c-8b98ff1aa0a6" />

## Features

- **Capture:** Add new insights instantly from your terminal.
- **Review:** Spaced Repetition System (SRS) to ensure long-term retention.
- **TUI:** Interactive Terminal User Interface for easy management.
- **Stats:** Track your learning progress over time.

## Prerequisites
must have [Go](https://go.dev/doc/install) installed in your system at least 1.25.7

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
til delete  # Deletes an entry, all, or by tags
til edit    # Edits an entry
til review  # Start a review session
til list    # List all entries
til stats   # Show learning statistics
til db      # current command is to show the database file path
```

## Notes
- this is just a simple personal project, it will definitely have a couple of undiscovered bugs.
- idk what else to say 


# SkypeHistoryViewer-go

A command-line tool to view and search Skype chat history from exported JSON files.

## Features

- 📱 **Command Line Interface**: Easy-to-use CLI with subcommands
- 🔍 **Advanced Search**: Search through messages with various filters
- 📊 **Statistics**: View detailed statistics about your chat history
- 💬 **Conversation Viewer**: Browse conversations with pagination
- 📎 **Export Functionality**: Export individual conversations to JSON
- 🎨 **Colored Output**: Beautiful colored terminal output
- ⚡ **Performance**: Optimized for large chat histories with progress indicators

## Installation

### From Source

```bash
git clone https://github.com/beckxie/SkypeHistoryViewer-go.git
cd SkypeHistoryViewer-go
go build -o skype-viewer
```

### Using Go Install

```bash
go install github.com/beckxie/SkypeHistoryViewer-go@latest
```

## Usage

### Basic Usage

```bash
# Show help
skype-viewer --help

# List all conversations
skype-viewer list -f /path/to/messages.json

# View a specific conversation
skype-viewer view 1 -f /path/to/messages.json

# Search through messages
skype-viewer search -q "search term" -f /path/to/messages.json

# Show statistics
skype-viewer stats -f /path/to/messages.json

# Export a conversation
skype-viewer export 1 -f /path/to/messages.json -o output.json
```

### Commands

#### `list` - List all conversations

```bash
skype-viewer list -f messages.json [flags]

Flags:
  --show-system    Include system messages in counts
```

#### `view` - View messages from a conversation

```bash
skype-viewer view [conversation-number] -f messages.json [flags]

Flags:
  --page-size int       Number of messages per page (default 20)
  --newest-first        Sort messages newest first
  --show-system         Show system messages
  --date-from string    Filter messages from this date (YYYY-MM-DD)
  --date-to string      Filter messages to this date (YYYY-MM-DD)
```

#### `search` - Search through messages

```bash
skype-viewer search -q "query" -f messages.json [flags]

Flags:
  -q, --query string         Search query text (required)
  --content                  Search in message content (default true)
  --sender                   Search in sender names (default true)
  --case-sensitive           Case-sensitive search
  --conversation string      Filter by conversation name
  --limit int                Maximum number of results (default 50)
  --date-from string         Search from this date (YYYY-MM-DD)
  --date-to string           Search to this date (YYYY-MM-DD)
```

#### `export` - Export a conversation

```bash
skype-viewer export [conversation-number] -f messages.json [flags]

Flags:
  -o, --output string    Output file path (default: auto-generated)
```

#### `stats` - Display statistics

```bash
skype-viewer stats -f messages.json
```

### Global Flags

```bash
-f, --file string    Path to Skype export JSON file or directory
-v, --verbose        Enable verbose output
```

## Exporting Skype Data

To export your Skype chat history:

1. Visit [Skype Export Support](https://support.microsoft.com/en-us/skype/how-do-i-export-or-delete-my-skype-data-84546e00-2fef-4c45-8ef6-3a27f83242cc)
2. Sign in with your Microsoft Account
3. Request your export
4. Download the `messages.json` file when ready

## Examples

### Search for messages from a specific person

```bash
skype-viewer search -q "John" --sender -f messages.json
```

### View conversations with date filter

```bash
skype-viewer view 1 -f messages.json --date-from 2024-01-01 --date-to 2024-12-31
```

### Export conversation with custom output

```bash
skype-viewer export 5 -f messages.json -o "john_doe_chat.json"
```

### Interactive conversation viewing

```bash
# Without conversation number, enters interactive mode
skype-viewer view -f messages.json
```

## Features in Detail

- **Pagination**: Large conversations are paginated for easy navigation
- **Date Filtering**: Filter messages by date range
- **System Messages**: Option to show/hide system messages
- **Progress Indicators**: Visual progress for file loading and searching
- **Cache**: Search results are cached for faster repeated searches
- **Unicode Support**: Proper handling of emojis and special characters

## Requirements

- Go 1.21 or later
- Terminal with color support (for best experience)

## License

MIT License

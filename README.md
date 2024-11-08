# Gosh

**Oh my GoSh all my  scripts  in one binary**

**Still in developement...**
## Features
### Session Management
- **Sessionizer** (`gosh s` or `gosh sessionizer`): Manipulate tmux sessions with various subcommands:
  - `fs`: Connect to SSH hosts from your config
  - `fd`: Find and create tmux sessions for directories
  - `vf`: Quick-open directories in your editor within tmux
  - `fg`: Clone and set up GitHub repositories with tmux sessions

### File Operations
- **Cat** (`gosh cat` or `gosh c`): View file contents with syntax highlighting
- **Edit** (`gosh edit` or `gosh e`): Quick access to vim/vi editor

### GitHub Integration
- **Install** (`gosh install`): Install binaries from GitHub releases
  ```bash
  # Install specific repositories
  gosh install mikefarah/yq DnFreddie/gosh
  
  # Install to custom directory
  gosh install --target ~/.local/bin mikefarah/yq
  
  # Install with custom binary name
  gosh install cli/cli:gh
  
  # Install predefined toolbox
  gosh install --toolbox
  ```

### Snippets
- **Snippets** (`gosh snip`): Manage and use code snippets

## Installation

You can install gosh using the following methods:

1. Using go install:
```bash
go install github.com/DnFreddie/gosh@latest
```

2. Clone and build from source:
```bash
git clone https://github.com/DnFreddie/gosh.git
cd gosh
go build
```

## Configuration

### Shell Completion
Generate shell completions using:
```bash
gosh compl [command-name] --shell [bash|zsh|fish] --completion-dir ~/.local/share/completions
```

### Default Directories
- Binaries are installed to `~/.local/bin` by default
- Temporary files are stored in the system's temp directory
- Completions are stored in `~/.local/share/completions`

## Dependencies
- tmux
- vim/vi (for edit command)
- git (for GitHub operations)


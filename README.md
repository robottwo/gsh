# gsh

[![License](https://img.shields.io/github/license/atinylittleshell/gsh.svg)](https://github.com/atinylittleshell/gsh/blob/main/LICENSE)
[![Release](https://img.shields.io/github/release/atinylittleshell/gsh.svg)](https://github.com/atinylittleshell/gsh/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/atinylittleshell/gsh/ci.yml?branch=main)](https://github.com/atinylittleshell/gsh/actions)
[![Test Coverage](https://codecov.io/gh/atinylittleshell/gsh/branch/main/graph/badge.svg?token=U7HWPOEPTF)](https://codecov.io/gh/atinylittleshell/gsh)

<p align="center">
A Modern, POSIX-compatible, <ins>G</ins>enerative <ins>Sh</ins>ell.
</p>

## Status

This project is in early development stage. Use at your own risk!
Please expect bugs, incomplete features and breaking changes.

That said, if you can try it out and provide feedback,
that would help make gsh more useful!

## Goals

- **POSIX-compatibility**: you shouldn't need to learn a new shell language to adopt gsh
- **Embrace AI**: gsh is designed from the ground up for the generative AI era to provide you with intelligent assistance at the right place, right time
- **Privacy**: gsh should allow you to use local LLMs for guaranteed privacy
- **Customizability**: gsh is _your_ shell. You should be able to configure it to your liking
- **Extensibility**: gsh should allow the community to build and share extensions to make it more useful

## But what does being "generative" mean?

### Generative suggestion of shell commands

gsh will automatically suggest the next command you are likely want to run.

![Generative Suggestion](assets/prediction.gif)

### Command explanation

gsh will provide an explanation of the command you are about to run.

![Command Explanation](assets/explanation.gif)

### Agent

gsh can act as an agent that invoke commands on your behalf.
Commands starting with "#" are sent to the agent as a chat message.

![Agent](assets/agent.gif)

gsh can even code for you!

![Agent Coding](assets/agent_coding.gif)

#### Chat Macros

You can use chat macros to quickly send frequently used messages to the agent.

A chat macro starts with "#/" followed by the macro name. The default configuration comes with a few chat macros:

```bash
# Summarize git changes
gsh> #/gitdiff

# Commit and push changes
gsh> #/gitpush

# Review changes and get suggestions
gsh> #/gitreview
```

You can customize your own macros by modifying the `GSH_AGENT_MACROS` configuration in your `.gshrc` file.
The value should be a JSON object mapping macro names to their corresponding chat messages.
See [Configuration](#configuration) for more details.

#### Agent Controls

Agent controls are built-in commands that help you manage your interaction with the agent.
An agent control starts with "#!" followed by the control name.

Currently supported controls:
```bash
# Reset the current chat session and start fresh
gsh> #!new

# Show token usage statistics for the current chat session
gsh> #!tokens
```

### Supports both local and remote LLMs

gsh can run with either

- Local LLMs through [Ollama](https://ollama.com/)
- Or remote LLMs through an OpenAI API-compatible endpoint, such as [OpenRouter](https://openrouter.ai/)

## Installation

To install gsh:

```bash
# Linux and macOS through Homebrew
brew tap atinylittleshell/gsh https://github.com/atinylittleshell/gsh
brew install atinylittleshell/gsh/gsh

# You can use gsh on arch, btw
yay -S gsh-bin
```

Windows is not supported (yet).

### Upgrading

gsh can automatically detect newer versions and self update.

## Building from Source

To build gsh from source, ensure you have Go installed and run the following command:

```bash
make build
```

This will compile the project and place the binary in the `./bin` directory.

## Configuration

gsh can be configured through a configuration file located at `~/.gshrc`.
Configuration options and default values can be found in [.gshrc.default](./cmd/gsh/.gshrc.default).

gsh also loads a `~/.gshenv` file, right after loading `~/.gshrc`.
This file can be used to set environment variables that the gsh session will use.

When launched as a login shell (`gsh -l`),
gsh will also load `/etc/profile` and `~/.gsh_profile` at start (before `~/.gshrc`).

### Custom command prompt

You can use [Starship.rs](https://starship.rs/) to render a custom command line prompt.
See [.gshrc.starship](./cmd/gsh/.gshrc.starship) for an example configuration.

## Usage

### Manually

You can manually start gsh from an existing shell:

```bash
gsh
```

### Automatically, through an existing shell

You can also automatically launch gsh from another shell's configuration file:

```bash
# For bash
echo "gsh" | tee -a ~/.bashrc
```

```bash
# For zsh
echo "gsh" | tee -a ~/.zshrc

# Your zsh config may have set "gsh" as an alias for `git show`.
# In that case, you would need to use the full path to gsh.
echo "/full/path/to/gsh" | tee -a ~/.zshrc
```

### Automatically, as your default shell

Or, you can set gsh as your default shell.
This is not recommended at the moment as gsh is still in early development.
But if you know what you are doing, you can do so by:

```bash
# Get the absolute path to gsh by running `which gsh`
which gsh

# Add gsh to the list of approved shells
echo "/path/to/gsh" | sudo tee -a /etc/shells

# Change your default shell to gsh
chsh -s "/path/to/gsh"
```

## Default Key Bindings

gsh provides a set of default key bindings for navigating and editing text input.
These key bindings are designed to be familiar to users of traditional shells and text editors.
It's on the roadmap to allow users to customize these key bindings.

- **Character Forward**: `Right Arrow`, `Ctrl+F`
- **Character Backward**: `Left Arrow`, `Ctrl+B`
- **Word Forward**: `Alt+Right Arrow`, `Ctrl+Right Arrow`, `Alt+F`
- **Word Backward**: `Alt+Left Arrow`, `Ctrl+Left Arrow`, `Alt+B`
- **Delete Word Backward**: `Alt+Backspace`, `Ctrl+W`
- **Delete Word Forward**: `Alt+Delete`, `Alt+D`
- **Delete After Cursor**: `Ctrl+K`
- **Delete Before Cursor**: `Ctrl+U`
- **Delete Character Backward**: `Backspace`, `Ctrl+H`
- **Delete Character Forward**: `Delete`, `Ctrl+D`
- **Line Start**: `Home`, `Ctrl+A`
- **Line End**: `End`, `Ctrl+E`
- **Paste**: `Ctrl+V`

## Roadmap

See [ROADMAP.md](./ROADMAP.md) for what's already planned.
Feel free to suggest new features by opening an issue!

## Acknowledgements

gsh is built on top of many great open source projects. Most notably:

- [mvdan/sh](https://github.com/mvdan/sh) - A shell parser, formatter, and interpreter
- [bubbletea](https://github.com/charmbracelet/bubbletea) - A powerful little TUI framework
- [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go
- [gorm](https://github.com/go-gorm/gorm) - The fantastic ORM library for Golang
- [go-openai](https://github.com/sashabaranov/go-openai) - A Go client for the OpenAI API

## Support This Project

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/onelittleshell)

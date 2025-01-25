# Bash Parity

- history expansion
- support bash completion
- multiline input

# Ergonomics

- keybind configuration (inputrc)
- syntax highlighting in shellinput
- better rendering of agent loading state

# AI

- agent chat macros
  - ui auto suggestions
- stop using go-openai - use raw http requests
- support custom instructions
  - use readme as agent context
- explicitly specify directory when runnning subshell commands (e.g. git status retriever)
- [set ollama context window](https://github.com/ollama/ollama/pull/6504)
- built-in eval
- MCP support
  - allow agent to search the web
  - allow agent to browse web urls
- persist agent message history
- training a small command prediction model

# Engineering

- limit total history size

# Distribution

- Official Homebrew
- apt
- nix
- Starship
- Windows

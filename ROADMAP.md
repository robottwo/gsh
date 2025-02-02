# Bash Parity

- history expansion
- [1.0] multiline input

# Ergonomics

- keybind configuration (inputrc)
- syntax highlighting in shellinput
- better rendering of agent loading state

# AI

- agent chat macros
  - ui auto suggestions
- stop using go-openai - use raw http requests
- [1.0] support custom instructions
- support agent modifying memory
- [set ollama context window](https://github.com/ollama/ollama/pull/6504)
- [1.0] built-in eval
- MCP support
  - allow agent to search the web
  - allow agent to browse web urls
- training a small command prediction model
  - log context and prediction history

# Engineering

- limit total history size

# Distribution

- Official Homebrew
- apt
- nix
- Starship
- Windows

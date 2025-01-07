# Bash Parity

- allow selecting past commands in shellinput
- support bash completion
- multiline input

# Ergonomics

- keybind configuration (inputrc)
- syntax highlighting in shellinput
- better rendering of agent loading state
- allow sigint to cancel agent chat

# AI

- set ollama context window
- RAG with aliases
- make current directory more clear in concise history context
- support custom instructions
- built-in eval
- allow agent to search the web
- allow agent to browse web urls
- agent message pruning using LLM
- persist agent message history
- use readme as agent context
- training a small command prediction model

# Engineering

- support concurrent writes to output/error buffers
- pre-fetch context when rendering prompt to avoid competing with user commands
- limit total history size

# Distribution

- Official Homebrew
- Windows

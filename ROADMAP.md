# Bash Parity

- support bash completion
- multiline input

# Ergonomics

- allow "yes to all" for some agent tools
- keybind configuration (inputrc)
- syntax highlighting in shellinput
- better rendering of agent loading state
- allow sigint to cancel agent chat

# AI

- explicitly specify directory when runnning subshell commands (e.g. git status retriever)
- set ollama context window
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
- limit total history size

# Distribution

- Official Homebrew
- Windows

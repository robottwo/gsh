# Bash Parity

- history command
- support bash completion
- multiline input

# Ergonomics

- allow "yes to all" for some agent tools
- keybind configuration (inputrc)
- syntax highlighting in shellinput
- better rendering of agent loading state

# AI

- allowlist of agent commands
- stop using go-openai - use raw http requests
- leverage openrouter's middle-out transform and replace our own pruning
- support custom instructions
- use readme as agent context
- explicitly specify directory when runnning subshell commands (e.g. git status retriever)
- set ollama context window
- make current directory more clear in concise history context
- built-in eval
- allow agent to search the web
- allow agent to browse web urls
- persist agent message history
- training a small command prediction model

# Engineering

- support concurrent writes to output/error buffers
- limit total history size

# Distribution

- Official Homebrew
- Windows

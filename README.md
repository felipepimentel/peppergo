# PepperGo

A flexible and extensible multi-agent system implemented in Go, supporting multiple AI frameworks transparently.

## Features

- 🤖 Multi-agent system with dynamic capabilities
- 🔌 Support for multiple AI providers (OpenAI, Anthropic, etc.)
- 🛠️ Extensible tool system
- 📦 Easy-to-use agent configuration via YAML
- 🔒 Built-in security features
- 📊 Comprehensive monitoring and logging
- 🚀 High performance and concurrent safety

## Project Structure

```
peppergo/
├── cmd/                    # Main applications
│   └── peppergo/          # CLI application
├── internal/              # Private application code
│   ├── agent/            # Agent implementation
│   ├── capability/       # Agent capabilities
│   ├── provider/         # AI provider integrations
│   ├── tool/            # Agent tools
│   └── config/          # Configuration management
├── pkg/                  # Public libraries
│   ├── types/           # Shared types and interfaces
│   ├── log/             # Logging utilities
│   └── errors/          # Error definitions
├── api/                  # API definitions
│   └── proto/           # Protocol buffer definitions
├── assets/              # Project assets
│   ├── agents/          # Agent definitions
│   └── prompts/         # System prompts
├── scripts/             # Build and maintenance scripts
├── test/                # Integration tests
└── docs/                # Documentation
```

## Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler
- golangci-lint
- make

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/peppergo.git
   cd peppergo
   ```

2. Install development tools:
   ```bash
   make tools
   ```

3. Build the project:
   ```bash
   make build
   ```

## Quick Start

1. Create an agent configuration:
   ```yaml
   # assets/agents/example.yaml
   name: example-agent
   version: "1.0.0"
   description: "Example agent"
   
   capabilities:
     - basic_chat
     - code_review
   
   tools:
     - file_reader
     - code_analyzer
   ```

2. Use the agent in your code:
   ```go
   package main

   import (
       "context"
       "log"

       "github.com/yourusername/peppergo/pkg/agent"
       "github.com/yourusername/peppergo/pkg/provider/openai"
   )

   func main() {
       ctx := context.Background()

       // Create agent from configuration
       agent, err := agent.FromYAML("assets/agents/example.yaml")
       if err != nil {
           log.Fatal(err)
       }

       // Configure provider
       provider := openai.NewProvider(openai.Config{
           APIKey: "your-api-key",
       })
       agent.UseProvider(provider)

       // Initialize and use
       if err := agent.Initialize(ctx); err != nil {
           log.Fatal(err)
       }
       defer agent.Cleanup(ctx)

       response, err := agent.Execute(ctx, "Review this code for security issues")
       if err != nil {
           log.Fatal(err)
       }
       log.Printf("Response: %s", response)
   }
   ```

## Development

- Run tests: `make test`
- Run linter: `make lint`
- Generate docs: `make docs`
- Validate structure: `make validate`
- Run all checks: `make dev`

## Documentation

- [Agent System](docs/agents/README.md)
- [Provider Integration](docs/providers/README.md)
- [Capability System](docs/capabilities/README.md)
- [Tool System](docs/tools/README.md)
- [API Reference](docs/api/README.md)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Run tests and linting (`make dev`)
4. Commit your changes (`git commit -m 'Add amazing feature'`)
5. Push to the branch (`git push origin feature/amazing-feature`)
6. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to all contributors
- Inspired by various AI agent frameworks
- Built with ❤️ using Go
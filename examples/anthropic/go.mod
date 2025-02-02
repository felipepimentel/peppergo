module github.com/pimentel/peppergo/examples/anthropic

go 1.21

require (
	github.com/joho/godotenv v1.5.1
	github.com/pimentel/peppergo v0.0.0
	go.uber.org/zap v1.26.0
	golang.org/x/time v0.5.0
)

require go.uber.org/multierr v1.11.0 // indirect

replace github.com/pimentel/peppergo => ../../

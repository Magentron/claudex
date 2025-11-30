# Go Skill

<skill_expertise>
You are an expert in Go, with deep knowledge of idiomatic Go patterns and concurrency.
- **Simplicity**: Keep code simple and readable - Go favors clarity over cleverness
- **Concurrency**: Master goroutines, channels, and synchronization primitives
- **Error Handling**: Always handle errors explicitly - errors are values
- **Testing**: Use the built-in testing package effectively with table-driven tests
- **Composition**: Prefer composition over inheritance using interfaces and embedding
</skill_expertise>

<coding_standards>
- Follow Effective Go guidelines and Go Proverbs
- Use `gofmt` formatting (non-negotiable)
- Keep packages small and focused with clear responsibilities
- Use meaningful package names (avoid util, common, misc, helpers)
- Export only what needs to be exported (PascalCase for public, camelCase for private)
- Write self-documenting code with clear variable and function names
- Add godoc comments for all exported identifiers
- Keep functions small and focused on a single responsibility
</coding_standards>

<best_practices>
## Error Handling
- Always handle errors explicitly - never ignore with `_`
- Use `errors.Is` and `errors.As` for error checking (Go 1.13+)
- Wrap errors with context using `fmt.Errorf` with `%w`
- Create sentinel errors for expected error cases (`var ErrNotFound = errors.New("not found")`)
- Return early on errors (guard clauses)
- Don't panic - use panic only for truly unrecoverable errors
- Prefer returning errors over logging and continuing

## Concurrency
- Use goroutines for concurrent work
- Use channels for communication between goroutines
- Prefer `sync.WaitGroup` for waiting on multiple goroutines
- Use `context.Context` for cancellation and timeouts
- Avoid shared state; pass data explicitly through channels
- Use buffered channels when appropriate to avoid blocking
- Always ensure goroutines can exit to prevent leaks
- Use `sync.Once` for one-time initialization
- Protect shared state with `sync.Mutex` or `sync.RWMutex` when channels aren't suitable

## Code Organization
- Keep `main` package minimal - extract logic to other packages
- Use `internal` packages for private code not meant to be imported
- Group related types in the same file
- Define interfaces at the consumer site, not the producer (accept interfaces, return structs)
- Use embedding for interface composition and struct extension
- Organize imports into groups: standard library, third-party, internal
- Use package-level variables sparingly; prefer dependency injection

## Testing Patterns
- Use table-driven tests for comprehensive coverage
- Use subtests with `t.Run` for organized test output
- Use `testify/assert` or `testify/require` for assertions (optional but common)
- Use `httptest` package for HTTP handler testing
- Use `testing.Short()` to skip integration tests during quick runs
- Use `t.Cleanup()` for test cleanup instead of defer
- Mock interfaces for unit testing (use tools like `gomock` or `mockery`)
- Test exported behavior, not internal implementation

## Type Safety
- Use custom types for domain concepts (`type UserID string`)
- Prefer slices over arrays for flexibility
- Use pointers for optional struct fields and large structs
- Use value receivers unless you need to modify the receiver
- Use pointer receivers for consistency if any method needs it
- Be explicit about zero values and document when they're meaningful

## Performance
- Use `strings.Builder` for string concatenation in loops
- Pre-allocate slices when size is known (`make([]T, 0, capacity)`)
- Use `sync.Pool` for frequently allocated objects
- Avoid premature optimization - profile first
- Use `pprof` for CPU and memory profiling
- Use benchmarks to measure performance (`go test -bench`)
</best_practices>

<utils>
## Test Execution Commands
```bash
# Run all tests
cd {project_path} && go test ./... -v

# Run tests with race detector
cd {project_path} && go test -race ./...

# Run tests with coverage
cd {project_path} && go test -cover ./...

# Run tests with coverage report
cd {project_path} && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run specific test
cd {project_path} && go test -v -run TestFunctionName

# Run benchmarks
cd {project_path} && go test -bench=. -benchmem
```

## Quality Check Commands
- `go fmt ./...` - Format all code (or use gofmt)
- `go vet ./...` - Static analysis for common bugs
- `golangci-lint run` - Comprehensive linting (requires golangci-lint)
- `go mod tidy` - Clean up module dependencies
- `go mod verify` - Verify dependencies haven't been modified
- `staticcheck ./...` - Advanced static analysis (requires staticcheck)

## Go-Specific Commands
- `go build ./...` - Build all packages
- `go install ./...` - Build and install binaries
- `go generate ./...` - Run code generation
- `go list -m all` - List all module dependencies
- `go mod graph` - Print module dependency graph
- `go doc {package}` - View package documentation
- `go test -race ./...` - Run with race detector
- `go test -short ./...` - Skip long-running tests

## Debugging Commands
- `go run -race main.go` - Run with race detector
- `dlv debug` - Start Delve debugger (requires delve)
- `go build -gcflags="-m"` - Show compiler optimization decisions
</utils>

<mcp_tools>
- `mcp__context7__resolve-library-id` - Resolve library identifiers
- `mcp__context7__get-library-docs` - Get up-to-date library documentation
- `mcp__sequential-thinking__sequentialthinking` - Deep analysis for complex decisions
</mcp_tools>

<go_patterns>
## Functional Options Pattern
```go
type Config struct {
    timeout time.Duration
    retries int
}

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.timeout = d
    }
}

func NewClient(opts ...Option) *Client {
    cfg := &Config{
        timeout: 30 * time.Second,
        retries: 3,
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return &Client{config: cfg}
}
```

## Context Usage
```go
func DoWork(ctx context.Context) error {
    select {
    case <-time.After(2 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

## Table-Driven Tests
```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

## Interface Composition
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type ReadWriter interface {
    Reader
    Writer
}
```
</go_patterns>

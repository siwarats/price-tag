# price-tag

A web scraper written in Go for extracting product prices from web pages.

## Project Structure

```
price-tag/
├── cmd/
│   └── main.go          # Entry point
├── internal/
│   └── scraper/
│       └── scraper.go   # Scraping logic
├── pkg/
│   └── models/
│       └── price.go     # Data models
├── go.mod
└── README.md
```

## Requirements

- Go 1.23+

## Getting Started

```bash
# Clone the repository
git clone https://github.com/siwar/price-tag.git
cd price-tag

# Build
go build ./...

# Run
go run ./cmd/main.go
```

## Development

```bash
# Run tests
go test ./...

# Format code
go fmt ./...

# Lint
go vet ./...
```

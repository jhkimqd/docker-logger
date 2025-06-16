# go-docker-logger

## Overview
`go-docker-logger` is a log monitoring tool that captures Docker container logs from a selected Docker network and displays them in the terminal. The tool highlights warning and error messages for better visibility.

## Features
- Monitors logs from all containers in a specified Docker network
- Color-coded output for different log levels:
  - Green: Normal logs
  - Yellow: Warning messages
  - Red: Error messages
- Flexible log filtering options
- Service name filtering support
- Graceful handling of application termination

## Getting Started

### Prerequisites
- Go 1.16 or later
- Docker installed and running

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-docker-logger.git
   cd go-docker-logger
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

### Building
```bash
# Build for current platform
make build

# Build for multiple platforms
make build-all
```

### Usage

Basic usage with network name (required):
```bash
./docker-logger --network <network-name>
```

Available flags:
```bash
--network string    Docker network name to monitor (required)
--all              Show all logs (default: false)
--errors           Show error logs only
--warnings         Show warning logs only
--info             Show info logs only
--debug            Show debug logs only
--levels string    Comma-separated log levels (error,warn,info,debug)
--filter string    Additional keywords to filter, comma-separated
--service string   Filter logs by service name (partial match)
```

### Examples

Monitor all logs from a network:
```bash
./docker-logger --network my-network
```

Show only error and warning logs:
```bash
./docker-logger --network my-network --errors --warnings
```

Filter by log levels:
```bash
./docker-logger --network my-network --levels=error,warn,info
```

Filter by service name (partial match):
```bash
./docker-logger --network my-network --service api-service
```

Combine multiple filters:
```bash
./docker-logger --network my-network --service api --levels=error,warn --filter="database,auth"
```

Using make:
```bash
make run network=my-network
```

## Development

Run tests:
```bash
make test
```

Clean build artifacts:
```bash
make clean
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
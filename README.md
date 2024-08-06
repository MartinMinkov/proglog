# Distributed Log Service in Go

A robust, scalable, and secure distributed log service implemented in Go.

## Overview

This project implements a distributed log service that provides efficient storage and retrieval of log data across a distributed system. It leverages modern technologies and best practices to ensure high performance, security, and observability.

## Key Features

- **Distributed Log Storage**: Efficiently store and retrieve log data in a distributed environment.
- **High-Performance Architecture**:
  - Quick lookups using an index file
  - Efficient data storage with a separate store file
- **gRPC API**: Fast and efficient client-server communication.
- **Security**:
  - TLS encryption for secure connections
  - Authentication and Authorization using ACLs (Casbin)
- **Observability**: Integrated logging, metrics, and tracing for comprehensive system monitoring.

## Technology Stack

- **Go**: Primary programming language (version 1.21+)
- **Protocol Buffers & gRPC**: For efficient serialization and API definition
- **TLS**: For encrypted communications
- **Casbin**: For flexible, powerful access control
- **Makefile**: For streamlined build and management processes

## Getting Started

### Prerequisites

- Go 1.21+
- Protocol Buffers compiler
- gRPC tools
- Casbin

### Quick Start

1. Clone the repository:

```

git clone https://github.com/MArtinMinkov/proglog.git

```

2. Navigate to the project directory:

```

cd proglog

```

3. Install dependencies:

```

go mod download

```

4. Generate Protocol Buffers code and TLS certificates:

```

make compile
make gencert

```

5. Run the server:

```

make run

```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

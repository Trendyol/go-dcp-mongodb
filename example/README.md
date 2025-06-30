# Examples

This directory contains various examples demonstrating how to use the go-dcp-mongodb connector in different scenarios.

## Prerequisites

Before running any example, ensure you have:
- Go 1.19 or higher installed
- Docker and Docker Compose (for examples that require external services)
- Access to Couchbase and MongoDB instances (either local or containerized)

## Examples Overview

### 1. Default Mapper (`default-mapper/`)

The simplest example that uses the default mapper functionality.

**How to run:**

*Option 1: Direct Go execution*
```bash
cd example/default-mapper
go mod tidy
go run main.go
```

*Option 2: Docker*
```bash
# Build from project root
docker build -f example/default-mapper/Dockerfile -t go-dcp-mongodb-default-mapper .

# Run the container
docker run --rm --network host go-dcp-mongodb-default-mapper
```

### 2. Simple (`simple/`)

Demonstrates custom mapper implementation with detailed event processing.

**How to run:**

*Option 1: Direct Go execution*
```bash
cd example/simple
go mod tidy
go run main.go
```

*Option 2: Docker*
```bash
# Build from project root
docker build -f example/simple/Dockerfile -t go-dcp-mongodb-simple .

# Run the container
docker run --rm --network host go-dcp-mongodb-simple
```

### 3. Simple Logger (`simple-logger/`)

Shows how to integrate custom logging with the connector.

**How to run:**

*Option 1: Direct Go execution*
```bash
cd example/simple-logger
go mod tidy
go run main.go
```

*Option 2: Docker*
```bash
# Build from project root
docker build -f example/simple-logger/Dockerfile -t go-dcp-mongodb-simple-logger .

# Run the container
docker run --rm --network host go-dcp-mongodb-simple-logger
```

### 4. Struct Config (`struct-config/`)

Demonstrates programmatic configuration using Go structs instead of YAML files.

**How to run:**

*Option 1: Direct Go execution*
```bash
cd example/struct-config
go mod tidy
go run main.go
```

*Option 2: Docker*
```bash
# Build from project root
docker build -f example/struct-config/Dockerfile -t go-dcp-mongodb-struct-config .

# Run the container
docker run --rm --network host go-dcp-mongodb-struct-config
```

### 5. Grafana (`grafana/`)

Complete monitoring setup with Grafana, Prometheus, and automatic data seeding.

**How to run:**
```bash
cd example/grafana
docker-compose up --build
```

**Access URLs:**
- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090
- Couchbase Admin: http://localhost:8091
- MongoDB: localhost:27017
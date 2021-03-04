# Spacemesh Watch

A CLI tool for monitoring various protocol features of spacemesh nodes and alerts if assertions are detected.

## Prerequisites

The CLI requires Golang version `1.55` or above to be installed

## Installing

Run the following commands to install the CLI

```
go install github.com/spacemeshos/spacemesh-watch@latest
yes | cp $(go env GOPATH)/bin/spacemesh-watch /usr/local/bin
```

### Usage

To start monitoring nodes use the below command:

```
spacemesh-watch --nodes=localhost:8001,localhost:8002 
```

Here the GRPC service of two nodes are exposed on port 8001 and 8002 respectively. To find out above other CLI options run:

```
spacemesh-watch --help
```

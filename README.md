# Caching-Proxy

## Description

A caching proxy in Go that forwards requests to the target server and caches responses for GET requests.

## Functionality

- Accepts HTTP requests on the specified port.

- Forwards requests to the target server specified in the arguments.

- Caches responses for GET requests.

- Returns cached responses for repeated requests.

- Supports background clearing of stale data from the cache.

## Installation

- Clone the repository.

- Run `go mod tidy` to install dependencies.

- Run `./build.sh` to build the executable.

## Usage

Start the proxy server by specifying the port and target server:

    ./caching-proxy --port 8080 --origin https://api.example.com

### Parameters: <br>
`--port` : Port on which the proxy will work (for example, 8080).

`--origin` : Destination server URL (for example, https://api.example.com).
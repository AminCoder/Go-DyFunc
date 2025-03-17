# Go-DyFunc
 
`Go-DyFunc` is an open source Golang project that enables dynamic remote function invocation over HTTP. It provides a flexible registry to add and call multiple functions remotely while reducing network overhead. By leveraging Go’s built-in concurrency, DyFunc ensures efficient execution through goroutines.

## Features

- **Basic Authentication**  
  Secure your endpoints using HTTP Basic Auth to restrict unauthorized access.

- **Multi-Function Invocation**  
  Register and call several functions with a single HTTP request, which helps reduce network traffic and overall overhead.

- **Goroutine Utilization**  
  Take advantage of Go’s goroutines for concurrent execution, enhancing the performance of your remote function calls.

- **Middleware Support**  
  Easily integrate middleware to log or process HTTP request data before function execution.

## Getting Started

### Prerequisites

- Go 1.x or later
- Basic understanding of Golang and HTTP servers

### Installation


```bash
git get github.com/AminCoder/Go-DyFunc
```

## Usage
Below is a sample code snippet that demonstrates how to set up Go-DyFunc with several useful functions:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"
    Registry "github.com/AminCoder/Go-DyFunc/pkg/registry"
	"github.com/AminCoder/Go-DyFunc/pkg/server"
)

func main() {
	// Create a new registry for registering functions
	registry := Registry.New_Registry()

	// Register functions

	// Sum two integers
	registry.Add("sum", func(x int, y int) int {
		return x + y
	})

	// Reverse a string
	registry.Add("reverse", func(input string) string {
		var reversed string
		for _, r := range input {
			reversed = string(r) + reversed
		}
		return reversed
	})

	// Return the current server time
	registry.Add("now", func() string {
		return time.Now().Format(time.RFC3339)
	})

	// Divide two integers with error checking for division by zero
	registry.Add("divide", func(x int, y int) (float64, error) {
		if y == 0 {
			return 0, fmt.Errorf("division by zero is not allowed")
		}
		return float64(x) / float64(y), nil
	})

	// Add middleware to log request details
	registry.Use(func(data Registry.Entry_Data) error {
		fmt.Println("Request from:", data.Http_Request.RemoteAddr)
		return nil
	})

	// Set up Basic Authentication for the HTTP server
	server.Set_Basic_Auth("admin", "123")

	// Run the HTTP server on port 5001 with endpoint "/call-remote"
	server.Run_HTTP_Server(":5001", "/call-remote", registry)
}
```

### Example API Request


Once the server is running, you can invoke the functions remotely by sending an HTTP POST request to the /call-remote endpoint with the necessary parameters.

For example, to call the sum function:


``` bash
curl -X POST http://localhost:5001/call-remote -d '{"function":"sum", "args":[4,5]}'
```





## Contributing

Contributions are highly welcome! 
If you have suggestions, improvements, or bug fixes, please feel free to
 open an issue or submit a pull request.


## License
This project is licensed under the MIT License. See the LICENSE file for more details.


Contact
For questions or feedback, please contact `amin@yodevs.com`.


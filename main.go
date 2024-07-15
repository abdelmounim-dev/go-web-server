package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

// Custom Request struct
type Request struct {
	Method string
	Path   string
	Header map[string]string
	Body   string // Added to hold the request body
}

// Custom Response struct
type Response struct {
	StatusCode int
	Status     string
	Header     map[string]string
	Body       string
}

// Function to read a Request from a TCP connection
func readRequest(conn net.Conn) (*Request, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// Parse the request line
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid request line")
	}

	method := parts[0]
	path := parts[1]

	header := make(map[string]string)

	// Read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" {
			break // End of headers
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			header[headerParts[0]] = strings.TrimSpace(headerParts[1])
		}
	}

	// Read the body if it's a POST request or has a body
	var body string
	if method == "POST" {
		contentLength := header["Content-Length"]
		if contentLength != "" {
			length, err := strconv.Atoi(contentLength)
			if err == nil && length > 0 {
				bodyBytes := make([]byte, length)
				_, err = io.ReadFull(reader, bodyBytes)
				if err == nil {
					body = string(bodyBytes)
				}
			}
		}
	}

	return &Request{Method: method, Path: path, Header: header, Body: body}, nil
}

// Function to create a Response
func createResponse(statusCode int, body string) *Response {
	status := "OK"
	if statusCode != 200 {
		status = "Error"
	}
	return &Response{
		StatusCode: statusCode,
		Status:     status,
		Header:     map[string]string{"Content-Type": "text/plain"},
		Body:       body,
	}
}

// Function to write a Response to a TCP connection
func writeResponse(conn net.Conn, resp *Response) error {
	header := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.Status)
	for key, value := range resp.Header {
		header += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	header += "\r\n" + resp.Body

	_, err := conn.Write([]byte(header))
	return err
}

// Handle individual connections
func handleConnection(conn net.Conn) {
	defer conn.Close()
	req, err := readRequest(conn)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	fmt.Printf("Received request: %s %s\n", req.Method, req.Path)
	if req.Method == "POST" {
		fmt.Printf("Request Body: %s\n", req.Body)
	}

	// Create a response
	resp := createResponse(200, "Hello, World!")
	err = writeResponse(conn, resp)
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// Main function to start the server
func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

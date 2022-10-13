package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	responseProto = "HTTP/1.1"

	statusOK         		= 200
	statusMethodNotAllowed  = 405
)

var statusText = map[int]string {
	statusOK:         		"OK",
	statusMethodNotAllowed: "Method Not Allowed",
}

type Server struct {
	// Addr ("host:port") : specifies the TCP address of the server
	Addr string
	// DocRoot the root folder under which clients can potentially look up information.
	// Anything outside this should be "out-of-bounds"
	DocRoot string
}

type Request struct {
	Method string // e.g. "GET"
}

type Response struct {
	StatusCode int    // e.g. 200 / 405
	Proto string	  // HTTP/1.1
	FilePath string		  // For this application, we will hard-code this to whatever contents are available in "hello-world.txt"
}

func (s *Server) ListenAndServe() error {
	// Validate the configuration of the server
	if err := s.ValidateServerSetup(); err != nil {
		return fmt.Errorf("server is not setup correctly %v", err)
	}
	fmt.Println("Server setup valid!")

	// server should now start to listen on the configured address
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on", ln.Addr())

	// making sure the listener is closed when we exit
	defer func() {
		err = ln.Close()
		if err != nil {
			fmt.Println("error in closing listener", err)
		}
	}()

	// accept connections forever
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		fmt.Println("accepted connection", conn.RemoteAddr())
		go s.HandleConnection(conn)
	}
}

func (s *Server) ValidateServerSetup() error {
	// Validating the doc root of the server
	fi, err := os.Stat(s.DocRoot)

	if os.IsNotExist(err) {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("doc root %q is not a directory", s.DocRoot)
	}

	return nil
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	br := bufio.NewReader(conn)
	for {
		// Set timeout
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			log.Printf("Failed to set timeout for connection %v", conn)
			_ = conn.Close()
			return
		}

		// Read next request from the client
		req, err := ReadRequest(br)

		// Handle EOF
		if errors.Is(err, io.EOF) {
			log.Printf("Connection closed by %v", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// timeout in this application means we just close the connection
		// Note : proj3 might require you to do a bit more here
		if err, ok := err.(net.Error); ok && err.Timeout() {
			log.Printf("Connection to %v timed out", conn.RemoteAddr())
			_ = conn.Close()
			return
		}

		// Handle the request which is not a GET and immediately close the connection and return
		if err != nil {
			log.Printf("Handle bad request for error: %v", err)
			res := &Response{}
			res.HandleBadRequest()
			_ = res.Write(conn)
			_ = conn.Close()
			return
		}

		// Handle good request
		log.Printf("Handle good request: %v", req)
		res := s.HandleGoodRequest()
		err = res.Write(conn)
		if err != nil {
			fmt.Println(err)
		}

		// We'll never close the connection and handle as many requests for this connection and pass on this
		// responsibility to the timeout mechanism
	}
}

func (s *Server) HandleGoodRequest() (res *Response) {
	res = &Response{}
	res.HandleOK()
	res.FilePath = filepath.Join(s.DocRoot, "hello-world.txt")

	return res
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK() {
	res.init()
	res.StatusCode = statusOK
}

// HandleBadRequest prepares res to be a 405 Method Not allowed response
func (res *Response) HandleBadRequest() {
	res.init()
	res.StatusCode = statusMethodNotAllowed
	res.FilePath = ""
}

func (res *Response) init() {
	res.Proto = responseProto
}

func ReadRequest(br *bufio.Reader) (req *Request, err error) {
	req = &Request{}

	// Read start line
	line, err := ReadLine(br)
	if err != nil {
		return nil, err
	}

	req.Method, err = parseRequestLine(line)
	if err != nil {
		return nil, badStringError("malformed start line", line)
	}

	if !validMethod(req.Method) {
		return nil, badStringError("invalid method", req.Method)
	}

	for {
		line, err := ReadLine(br)
		if err != nil {
			return nil, err
		}
		if line == "" {
			// This marks header end
			break
		}
		fmt.Println("Read line from request", line)
	}

	return req, nil
}

// parseRequestLine parses "GET /foo HTTP/1.1" into its individual parts.
func parseRequestLine(line string) (string, error) {
	fields := strings.SplitN(line, " ", 2)
	if len(fields) != 2 {
		return "", fmt.Errorf("could not parse the request line, got fields %v", fields)
	}
	return fields[0], nil
}

func validMethod(method string) bool {
	return method == "GET"
}

func badStringError(what, val string) error {
	return fmt.Errorf("%s %q", what, val)
}

func (res *Response) Write(w io.Writer) error {
	bw := bufio.NewWriter(w)

	statusLine := fmt.Sprintf("%v %v %v\r\n", res.Proto, res.StatusCode, statusText[res.StatusCode])
	if _, err := bw.WriteString(statusLine); err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}
package server

import (
	"bufio"
	"strings"
	"testing"
)

func TestHandleConnection_Simple_GET(t *testing.T) {
	request := "GET /index.html HTTP/1.1\r\n" +
			   "Host: test\r\n" +
			   "\r\n"

	req, err := ReadRequest(bufio.NewReader(strings.NewReader(request)))
	if req.Method != "GET" || err != nil {
		t.Fatalf("incorrect parsing of request %v : %v", req, err)
	}
}

func TestHandleConnection_Simple_POST(t *testing.T) {
	request := "POST /index.html HTTP/1.1\r\n" +
		"Host: test\r\n" +
		"\r\n"

	req, err := ReadRequest(bufio.NewReader(strings.NewReader(request)))
	if req != nil || err == nil {
		t.Fatalf("POST should not be allowed, looks like it is! %v : %v", req, err)
	}
}
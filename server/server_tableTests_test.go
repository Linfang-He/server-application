package server

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func checkGoodRequest(t *testing.T, readErr error, reqGot, reqWant *Request) {
	if readErr != nil {
		t.Fatal(readErr)
	}
	// DeepEqual is a recursive relaxation of the == operator in go
	if !reflect.DeepEqual(*reqGot, *reqWant) {
		t.Fatalf("\ngot: %v\nwant: %v", reqGot, reqWant)
	}
}

func checkBadRequest(t *testing.T, readErr error, reqGot *Request) {
	if readErr == nil {
		t.Errorf("\ngot unexpected request: %v\nwant: error", reqGot)
	}
}

func TestReadGoodRequest(t *testing.T) {
	var tests = []struct {
		name    string
		reqText string
		reqWant *Request
	}{
		{
			"Simple GET",
			"GET /index.html HTTP/1.1\r\n" +
				"Host: test\r\n" +
				"\r\n",
			&Request{
				Method: "GET",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqGot, err := ReadRequest(bufio.NewReader(strings.NewReader(tt.reqText)))
			checkGoodRequest(t, err, reqGot, tt.reqWant)
		})
	}
}

func TestReadBadRequest(t *testing.T) {
	var tests = []struct {
		name string
		req  string
	}{
		{
			"Basic",
			"This is a bad request\r\n",
		},
		{
			"Empty",
			"\r\n",
		},
		{
			"InvalidHTTPVerb",
			"GETT /index.html HTTP/1.1\r\nHost: test\r\n\r\n",
		},
		{
			"NotSupportedHTTPVerb",
			"POST /index.html HTTP/1.0\r\nHost: test\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqGot, err := ReadRequest(bufio.NewReader(strings.NewReader(tt.req)))
			checkBadRequest(t, err, reqGot)
		})
	}
}

func TestReadMultipleRequests(t *testing.T) {
	var tests = []struct {
		name     string
		reqText  string
		reqsWant []*Request
	}{
		{
			"GoodGood",
			"GET /index.html HTTP/1.1\r\nHost: test\r\n\r\n" +
				"GET /index.html HTTP/1.1\r\nHost: test\r\n\r\n",
			[]*Request{
				{
					Method: "GET",
				},
				{
					Method: "GET",
				},
			},
		},
		{
			"GoodBad",
			"GET /index.html HTTP/1.1\r\nHost: test\r\n\r\n" +
				"GETT /index.html HTTP/1.1\r\nHost: test\r\n\r\n",
			[]*Request{
				{
					Method: "GET",
				},
				nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := bufio.NewReader(strings.NewReader(tt.reqText))
			for _, reqWant := range tt.reqsWant {
				reqGot, err := ReadRequest(br)
				if reqWant != nil {
					checkGoodRequest(t, err, reqGot, reqWant)
				} else {
					checkBadRequest(t, err, reqGot)
				}
			}
		})
	}
}
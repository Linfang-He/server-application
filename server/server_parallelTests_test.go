package server

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func checkBadRequest_slow(t *testing.T, readErr error, reqGot *Request) {
	if readErr == nil {
		t.Errorf("\ngot unexpected request: %v\nwant: error", reqGot)
	}
	time.Sleep(time.Second * 2)
}

func TestReadBadRequest_parallel(t *testing.T) {
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
			t.Parallel()
			reqGot, err := ReadRequest(bufio.NewReader(strings.NewReader(tt.req)))
			checkBadRequest_slow(t, err, reqGot)
		})
	}
}
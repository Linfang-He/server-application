package server

import (
	"bufio"
	"strings"
)

// ReadLine reads a single line ending with "\r\n" from br,
// striping the "\r\n" line end from the returned string.
// If any error occurs, data read before the error is also returned.
// You might find this function useful in parsing requests.
func ReadLine(br *bufio.Reader) (string, error) {
	var line string
	for {
		s, err := br.ReadString('\n')
		line += s
		// Return the error
		if err != nil {
			return line, err
		}
		// Return the line when reaching line end
		if strings.HasSuffix(line, "\r\n") {
			// Striping the line end
			line = line[:len(line)-2]
			return line, nil
		}
	}
}

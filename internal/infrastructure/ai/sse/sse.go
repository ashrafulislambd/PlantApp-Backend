// Package sse reads Server-Sent Events streams, the wire format both Gemini
// (alt=sse) and Groq (OpenAI-compatible stream:true) use for incremental
// replies.
package sse

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// ErrStop can be returned from a Read callback to end the stream early
// without Read reporting an error.
var ErrStop = errors.New("sse: stop")

// Read parses r as an SSE stream and calls onData with the data payload of
// each event (multi-line data fields are joined with "\n"). Comment lines
// (": ...") and fields other than "data" are ignored. It returns nil at EOF
// or when onData returns ErrStop, and any other error otherwise.
//
// A bufio.Reader is used instead of bufio.Scanner so a single very long
// data line cannot hit a token-size limit.
func Read(r io.Reader, onData func(data string) error) error {
	br := bufio.NewReader(r)
	var data []string

	flush := func() error {
		if len(data) == 0 {
			return nil
		}
		payload := strings.Join(data, "\n")
		data = data[:0]
		return onData(payload)
	}

	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimRight(line, "\r\n")
			switch {
			case line == "":
				if ferr := flush(); ferr != nil {
					return stopToNil(ferr)
				}
			case strings.HasPrefix(line, ":"):
				// comment / keep-alive
			case strings.HasPrefix(line, "data:"):
				data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
			}
		}
		if err != nil {
			if err == io.EOF {
				// The stream may end without a trailing blank line.
				return stopToNil(flush())
			}
			return err
		}
	}
}

func stopToNil(err error) error {
	if errors.Is(err, ErrStop) {
		return nil
	}
	return err
}

package sse

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

const dataPrefix = "data: "

// Parser extracts SSE data frames.
type Parser struct {
	scanner *bufio.Scanner
	buffer  bytes.Buffer
}

// NewParser wraps a reader.
func NewParser(r io.Reader) *Parser {
	return &Parser{scanner: bufio.NewScanner(r)}
}

// NextEvent returns the next data payload.
func (p *Parser) NextEvent() ([]byte, error) {
	p.buffer.Reset()

	for p.scanner.Scan() {
		line := p.scanner.Text()

		if line == "" {
			if p.buffer.Len() > 0 {
				result := make([]byte, p.buffer.Len())
				copy(result, p.buffer.Bytes())
				return result, nil
			}
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue
		}

		if strings.HasPrefix(line, dataPrefix) {
			data := strings.TrimPrefix(line, dataPrefix)
			if p.buffer.Len() > 0 {
				p.buffer.WriteByte('\n')
			}
			p.buffer.WriteString(data)
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, err
	}

	if p.buffer.Len() > 0 {
		result := make([]byte, p.buffer.Len())
		copy(result, p.buffer.Bytes())
		return result, io.EOF
	}

	return nil, io.EOF
}

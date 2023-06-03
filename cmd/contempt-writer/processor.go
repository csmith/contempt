package main

import "bytes"

// Processor is responsible for processing a contempt directive within a Dockerfile/Containerfile.
type Processor interface {
	// Prefix identifies the single word prefix that will be used to invoke this processor.
	// This is what the user will type in the Dockerfile/Containerfile, e.g. `#C: my-prefix some arguments`.
	Prefix() string

	// Take returns the length of any previously generated instructions at the start of the buffer.
	// Take may freely read from the given buffer. If there are no relevant instructions, 0 should be returned.
	Take(buf *bytes.Buffer) (int, error)

	// Write generates new output in the buffer using the given arguments.
	Write(args []string, buf *bytes.Buffer) error
}

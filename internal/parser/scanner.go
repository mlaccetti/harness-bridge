package parser

import (
	"bufio"
	"io"
)

// Claude/Codex JSONL lines can exceed bufio's default 64KiB max token (large tool payloads).
const maxJSONLLineBytes = 256 << 20 // 256 MiB

func newJSONLScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, maxJSONLLineBytes)
	return sc
}

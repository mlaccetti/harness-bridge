package cli

import (
	"bytes"
	"testing"
)

func TestRootCommand_listRequiresNoArgs(t *testing.T) {
	root := NewRoot()
	var buf bytes.Buffer
	WireOutput(root, &buf)
	root.SetArgs([]string{"list", "--from", "claude", "--limit", "0"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
}

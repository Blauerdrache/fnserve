package runtime

import (
	"bytes"
	"fmt"
	"os/exec"
)

type GoRuntime struct{}

func (r *GoRuntime) Execute(functionPath string, event []byte, ctx Context) ([]byte, error) {
	// Run the compiled Go binary directly
	cmd := exec.Command(functionPath)
	cmd.Stdin = bytes.NewReader(event)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go error: %s", out.String())
	}

	return out.Bytes(), nil
}

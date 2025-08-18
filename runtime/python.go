package runtime

import (
	"bytes"
	"fmt"
	"os/exec"
)

type PythonRuntime struct{}

func (r *PythonRuntime) Execute(functionPath string, event []byte, ctx Context) ([]byte, error) {
	// For now, just call python directly on the file
	cmd := exec.Command("python3", functionPath)
	cmd.Stdin = bytes.NewReader(event)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("python error: %s", out.String())
	}

	return out.Bytes(), nil
}

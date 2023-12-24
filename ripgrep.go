package preditor

import (
	"bytes"
	"fmt"
	"os/exec"
)

func runRipgrep(pattern string) chan [][]byte {
	output := make(chan [][]byte)
	go func() {
		cmd := exec.Command("rg", "--vimgrep", pattern)
		stdError := &bytes.Buffer{}
		stdOut := &bytes.Buffer{}

		cmd.Stderr = stdError
		cmd.Stdout = stdOut
		if err := cmd.Run(); err != nil {
			fmt.Println("ERROR running rg:", err.Error())
			return
		}

		output <- bytes.Split(stdOut.Bytes(), []byte("\n"))
	}()

	return output
}

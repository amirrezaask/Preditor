package preditor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func RipgrepAsync(pattern string, dir string) chan [][]byte {
	output := make(chan [][]byte)
	go func() {
		cmd := exec.Command("rg", "--vimgrep", pattern)
		stdError := &bytes.Buffer{}
		stdOut := &bytes.Buffer{}

		cmd.Stderr = stdError
		cmd.Stdout = stdOut
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			fmt.Println("ERROR running rg:", err.Error())
			return
		}

		output <- bytes.Split(stdOut.Bytes(), []byte("\n"))
	}()

	return output
}

func RipgrepFiles(cwd string) []string {
	cmd := exec.Command("rg", "--files")
	stdError := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}

	cmd.Stderr = stdError
	cmd.Stdout = stdOut
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		fmt.Println("ERROR running rg files:", err.Error())
		return nil
	}

	return strings.Split(stdOut.String(), "\n")
}

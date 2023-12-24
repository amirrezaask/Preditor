package preditor

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func NewCompilationBuffer(parent *Context, cfg *Config, command string) (*TextBuffer, error) {
	tb, err := NewTextBuffer(parent, cfg, "*Compilation*")
	if err != nil {
		return nil, err
	}
	cwd := parent.getCWD()

	tb.Readonly = true
	runCompileCommand := func() {
		tb.Content = nil
		tb.Content = append(tb.Content, []byte(fmt.Sprintf("Command: %s\n", command))...)
		tb.Content = append(tb.Content, []byte(fmt.Sprintf("Dir: %s\n", cwd))...)
		go func() {
			segs := strings.Split(command, " ")
			var args []string
			bin := segs[0]
			if len(segs) > 1 {
				args = segs[1:]
			}
			cmd := exec.Command(bin, args...)
			cmd.Dir = cwd
			since := time.Now()
			output, err := cmd.CombinedOutput()
			if err != nil {
				tb.Content = []byte(err.Error())
				tb.Content = append(tb.Content, '\n')
			}
			tb.Content = append(tb.Content, output...)
			tb.Content = append(tb.Content, []byte(fmt.Sprintf("Done in %s\n", time.Since(since)))...)

		}()

	}
	tb.keymaps[1].SetKeyCommand(Key{K: "g"}, MakeCommand(func(b *TextBuffer) error {
		runCompileCommand()

		return nil
	}))

	runCompileCommand()
	return tb, nil
}

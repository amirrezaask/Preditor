package preditor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func OpenLocationInCurrentLine(c *Context) error {
	b, ok := c.ActiveBuffer().(*Buffer)
	if !ok {
		return nil
	}

	line := BufferGetCurrentLine(b)
	if line == nil || len(line) < 1 {
		return nil
	}

	segs := bytes.SplitN(line, []byte(":"), 4)
	if len(segs) < 2 {
		return nil

	}

	var targetWindow *Window
	for _, col := range c.Windows {
		for _, win := range col {
			if c.ActiveWindowIndex != win.ID {
				targetWindow = win
				break
			}
		}
	}

	filename := segs[0]
	var lineNum int
	var col int
	var err error
	switch len(segs) {
	case 3:
		//filename:line: text
		lineNum, err = strconv.Atoi(string(segs[1]))
		if err != nil {
		}
	case 4:
		//filename:line:col: text
		lineNum, err = strconv.Atoi(string(segs[1]))
		if err != nil {
		}
		col, err = strconv.Atoi(string(segs[2]))
		if err != nil {
		}

	}
	_ = SwitchOrOpenFileInWindow(c, c.Cfg, string(filename), &Position{Line: lineNum, Column: col}, targetWindow)

	c.ActiveWindowIndex = targetWindow.ID
	return nil
}

func RunCommandWithOutputBuffer(parent *Context, cfg *Config, bufferName string, command string) (*Buffer, error) {
	tb, err := NewBuffer(parent, cfg, bufferName)
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

	tb.keymaps[1].SetKeyCommand(Key{K: "g"}, MakeCommand(func(b *Buffer) error {
		runCompileCommand()

		return nil
	}))
	tb.keymaps[1].SetKeyCommand(Key{K: "<enter>"}, OpenLocationInCurrentLine)

	runCompileCommand()
	return tb, nil
}

func NewGrepBuffer(parent *Context, cfg *Config, command string) (*Buffer, error) {
	return RunCommandWithOutputBuffer(parent, cfg, "*Grep", command)
}

func NewCompilationBuffer(parent *Context, cfg *Config, command string) (*Buffer, error) {
	return RunCommandWithOutputBuffer(parent, cfg, "*Compilation*", command)

}

package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
	"github.com/valyala/fasttemplate"
	"golang.org/x/sys/windows"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

type Command struct {
	Name     string        `json:"name,omitempty" toml:"name,omitempty" ini:"name,omitempty"`
	Commands []string      `json:"commands,omitempty" toml:"commands,omitempty" ini:"commands,omitempty"`
	Command  string        `json:"command,omitempty" toml:"command,omitempty" ini:"command,omitempty"`
	Charset  string        `json:"charset,omitempty" toml:"charset,omitempty" ini:"charset,omitempty"`
	Shell    string        `json:"shell,omitempty" toml:"shell,omitempty" ini:"shell,omitempty"`
	Dir      string        `json:"dir,omitempty" toml:"dir,omitempty" ini:"dir,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty" toml:"timeout,omitempty" ini:"timeout,omitempty"`
}

func Run(ctx context.Context, c Command, args map[string]string) (output []byte, err error) {
	commands := c.Commands
	if len(commands) == 0 && c.Command != "" {
		commands = append(commands, c.Command)
	}

	var outputs []string
	for _, cmd := range commands {
		out, e := ExecProgram(ctx, Program{
			Shell:   c.Shell,
			Command: cmd,
			Dir:     c.Dir,
			Charset: c.Charset,
			Timeout: c.Timeout,
		}, args)
		if out != "" {
			outputs = append(outputs, out)
		}
		if e != nil {
			err = e
			break
		}
	}

	output = []byte(strings.Join(outputs, "\n"))
	return
}

type Program struct {
	Shell   string
	Command string
	Dir     string
	Charset string
	Timeout time.Duration
}

func ExecProgram(ctx context.Context, prog Program, replArgs map[string]string) (out string, err error) {
	var charset encoding.Encoding
	if prog.Charset != "" {
		if charset, err = htmlindex.Get(prog.Charset); err != nil {
			return
		}
	}

	command := tplRepl(prog.Command, replArgs)

	var args []string

	if prog.Shell == "" {
		if args, err = shlex.Split(command); err != nil {
			return
		}

		if len(args) == 0 || (len(args) == 1 && args[0] == "") {
			err = errors.New("empty command")
			return
		}
	} else {
		args = append(args, prog.Shell)
	}

	if charset != nil {
		e := charset.NewEncoder()
		for i, arg := range args {
			if args[i], err = e.String(arg); err != nil {
				return
			}
		}

		if command, err = e.String(command); err != nil {
			return
		}
	}

	if prog.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, prog.Timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_NEW_PROCESS_GROUP}

	cmd.Dir = prog.Dir
	cmd.Env = os.Environ()

	if prog.Shell != "" {
		cmd.Stdin = strings.NewReader(command + "\n")
	}

	var output []byte
	output, err = cmd.CombinedOutput()
	if len(output) > 0 {
		if charset != nil {
			if output, err = charset.NewDecoder().Bytes(output); err != nil {
				return
			}
		}
		out = string(output)
	}
	return
}

func tplRepl(src string, args map[string]string) string {
	out, err := fasttemplate.New(src, "{{", "}}").ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
		sTag := strings.TrimSpace(tag)
		switch {
		case strings.HasPrefix(sTag, "args."):
			return w.Write([]byte(args[sTag[5:]]))
		case strings.HasPrefix(sTag, "env."):
			return w.Write([]byte(os.Getenv(sTag[4:])))
		}
		return w.Write([]byte("{{" + tag + "}}"))
	})
	if err != nil {
		return src
	}
	return out
}

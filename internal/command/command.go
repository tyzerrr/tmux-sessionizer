package command

import (
	"bytes"
	"fmt"
	"os/exec"
)

var (
	ErrTmuxCmdNoOutBuf = fmt.Errorf("tmux command has no output buffer")
)

type FzfCommand struct {
	*exec.Cmd
	inBuf  *bytes.Buffer
	outBuf *bytes.Buffer
}

func NewFzfCommand() *FzfCommand {
	cmd := exec.Command("fzf")
	inBuf, outBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdin, cmd.Stdout = inBuf, outBuf
	return &FzfCommand{
		Cmd:    cmd,
		inBuf:  inBuf,
		outBuf: outBuf,
	}
}

func (fc *FzfCommand) Run() error {
	if err := fc.Cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (fc *FzfCommand) InBuf() *bytes.Buffer {
	return fc.inBuf
}

func (fc *FzfCommand) OutBuf() *bytes.Buffer {
	return fc.outBuf
}

type TmuxCommand struct {
	*exec.Cmd
	outBuf *bytes.Buffer
}

func NewTmuxCommand(args ...string) *TmuxCommand {
	cmd := exec.Command("tmux", args...)
	ob := &bytes.Buffer{}
	cmd.Stdout = ob
	/*
	   When tmux try to attach, real tty is necessary.
	   But as default, exec.Command provides virtual in-memory pipe.
	   So tmux throws the error.
	*/
	return &TmuxCommand{
		Cmd:    cmd,
		outBuf: ob,
	}
}

func (tc *TmuxCommand) OutBuf() *bytes.Buffer {
	return tc.outBuf
}

func (tc *TmuxCommand) Run() error {
	if tc.Stdout == nil {
		return ErrTmuxCmdNoOutBuf
	}
	if err := tc.Cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (tc *TmuxCommand) Out() []byte {
	return tc.Stdout.(*bytes.Buffer).Bytes()
}

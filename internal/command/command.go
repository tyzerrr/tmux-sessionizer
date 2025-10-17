package command

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
)

var (
	ErrTmuxCmdNoOutBuf = errors.New("tmux command has no output buffer")
)

type FzfCommand struct {
	*exec.Cmd

	inBuf  *bytes.Buffer
	outBuf *bytes.Buffer
}

func NewFzfCommand(ctx context.Context) *FzfCommand {
	cmd := exec.CommandContext(ctx, "fzf")
	inBuf, outBuf := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdin, cmd.Stdout = inBuf, outBuf

	return &FzfCommand{
		Cmd:    cmd,
		inBuf:  inBuf,
		outBuf: outBuf,
	}
}

func (fc *FzfCommand) Run() error {
	err := fc.Cmd.Run()
	if err != nil {
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

func NewTmuxCommand(ctx context.Context, args ...string) *TmuxCommand {
	cmd := exec.CommandContext(ctx, "tmux", args...)
	ob := &bytes.Buffer{}
	/*
	   When tmux try to attach, real tty is necessary.
	   But as default, exec.CommandContext provides virtual in-memory pipe.
	   So tmux throws the error.
	*/
	cmd.Stdout = ob
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

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

	err := tc.Cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (tc *TmuxCommand) Out() []byte {
	buf, ok := tc.Stdout.(*bytes.Buffer)
	if !ok {
		return nil
	}

	return buf.Bytes()
}

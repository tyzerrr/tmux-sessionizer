package cmd

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/TlexCypher/my-tmux-sessionizer/handler"
	"github.com/urfave/cli/v3"
)

func newMockCmd() *cli.Command {
	return &cli.Command{
		Name:   "mock tmux-sessionizer",
		Usage:  "mock tmux session manager",
		Action: mockRun,
	}
}

func mockRun(ctx context.Context, cmd *cli.Command) error {
	mh := newMockSessionHandler()
	return runWithHandler(mh, ctx, cmd)
}

type MockSessionHandler struct{}

func newMockSessionHandler() handler.ISessionHandler {
	return &MockSessionHandler{}
}

func (mh *MockSessionHandler) NewSession(ctx context.Context) error {
	return nil
}

func (mh *MockSessionHandler) GrabExistingSession(ctx context.Context) error {
	return nil
}

func (mh *MockSessionHandler) CreateNewProjectSession(ctx context.Context) error {
	return nil
}

func (mh *MockSessionHandler) DeleteProjectSession(ctx context.Context) error {
	return nil
}

type args struct {
	cmd string
}

func TestRun(t *testing.T) {
	t.Parallel()

	var mu sync.RWMutex

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "normal case (tmux-sessionizer)",
			args: args{
				cmd: "",
			},
			wantErr: nil,
		},
		{
			name: "list case (tmux-sessionizer list)",
			args: args{
				cmd: "list",
			},
			wantErr: nil,
		},
		{
			name: "invalid command case",
			args: args{
				cmd: "invalid",
			},
			wantErr: ErrNoSuchCmd,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := newMockCmd()

			args := []string{"tmux-sessionizer"}
			if tt.args.cmd != "" {
				args = append(args, tt.args.cmd)
			}

			mu.Lock()

			err := cmd.Run(t.Context(), args)

			mu.Unlock()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

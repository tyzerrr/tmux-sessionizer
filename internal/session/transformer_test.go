package session

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTransformer_Transform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Replace dot to underscore",
			input: "/path/to/.config/tmux-sesionizer",
			want:  "/path/to/_config/tmux-sesionizer",
		},
		{
			name:  "Replace colon to underscore",
			input: "/path/to/project:session",
			want:  "/path/to/project;session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := NewTransformer().WithRule(
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ".", "_") },
					func(in string) string { return strings.ReplaceAll(in, "_", ".") },
				),
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ":", ";") },
					func(in string) string { return strings.ReplaceAll(in, ";", ":") },
				),
			)

			got := st.Transform(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("cannonicalizeSessionName() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTransformer_Revert(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Replace underscore back to dot",
			input: "/path/to/_config/tmux-sesionizer",
			want:  "/path/to/.config/tmux-sesionizer",
		},
		{
			name:  "Replace semicolon back to colon",
			input: "/path/to/project;session",
			want:  "/path/to/project:session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			st := NewTransformer().WithRule(
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ".", "_") },
					func(in string) string { return strings.ReplaceAll(in, "_", ".") },
				),
				NewTransformRule(
					func(in string) string { return strings.ReplaceAll(in, ":", ";") },
					func(in string) string { return strings.ReplaceAll(in, ";", ":") },
				),
			)

			got := st.Revert(tt.input)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Transformer.Revert() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

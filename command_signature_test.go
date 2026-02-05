package kong

import (
	"strings"
	"testing"
)

type signatureMethodCmd struct{}

func (signatureMethodCmd) Signature() string {
	return `name:"method" help:"method help" aliases:"m1,m2" hidden group:"Method Group"`
}

type signatureMethodRoot struct {
	Cmd *signatureMethodCmd `cmd:"" name:"tag" help:"tag help" aliases:"t1,t2" group:"Tag Group"`
}

func TestSignatureMethodOverridesTags(t *testing.T) {
	root := &signatureMethodRoot{Cmd: &signatureMethodCmd{}}
	k, err := New(root)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if len(k.Model.Node.Children) != 1 {
		t.Fatalf("expected 1 command, got %d", len(k.Model.Node.Children))
	}
	cmd := k.Model.Node.Children[0]
	if got, want := cmd.Name, "method"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	if got, want := strings.Join(cmd.Aliases, ","), "m1,m2"; got != want {
		t.Fatalf("aliases = %q, want %q", got, want)
	}
	if got, want := cmd.Help, "method help"; got != want {
		t.Fatalf("help = %q, want %q", got, want)
	}
	if !cmd.Hidden {
		t.Fatalf("hidden = %v, want true", cmd.Hidden)
	}
	if cmd.Group == nil || cmd.Group.Key != "Method Group" {
		key := ""
		if cmd.Group != nil {
			key = cmd.Group.Key
		}
		t.Fatalf("group key = %q, want %q", key, "Method Group")
	}
}

type signatureEmptyCmd struct{}

func (signatureEmptyCmd) Signature() string { return "" }

type signatureEmptyRoot struct {
	Cmd *signatureEmptyCmd `cmd:"" name:"tag" help:"tag help" aliases:"t1,t2" group:"Tag Group"`
}

func TestSignatureMethodEmptyUsesTags(t *testing.T) {
	root := &signatureEmptyRoot{Cmd: &signatureEmptyCmd{}}
	k, err := New(root)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	cmd := k.Model.Node.Children[0]
	if got, want := cmd.Name, "tag"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	if got, want := strings.Join(cmd.Aliases, ","), "t1,t2"; got != want {
		t.Fatalf("aliases = %q, want %q", got, want)
	}
	if got, want := cmd.Help, "tag help"; got != want {
		t.Fatalf("help = %q, want %q", got, want)
	}
	if cmd.Hidden {
		t.Fatalf("hidden = %v, want false", cmd.Hidden)
	}
	if cmd.Group == nil || cmd.Group.Key != "Tag Group" {
		key := ""
		if cmd.Group != nil {
			key = cmd.Group.Key
		}
		t.Fatalf("group key = %q, want %q", key, "Tag Group")
	}
}

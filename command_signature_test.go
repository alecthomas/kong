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

type signatureInheritedGroupCmd struct{}

func (signatureInheritedGroupCmd) Signature() string { return `help:"from signature"` }

type signatureInheritedGroupRoot struct {
	Nested struct {
		Cmd *signatureInheritedGroupCmd `cmd:""`
	} `embed:"" group:"Parent Group"`
}

type signatureMixedACmd struct{}

func (signatureMixedACmd) Signature() string { return `help:"sig-a-help" aliases:"sig-a"` }

type signatureMixedBCmd struct{}

func (signatureMixedBCmd) Signature() string { return `name:"sig-b" hidden` }

type signatureMixedRoot struct {
	Parent struct {
		A *signatureMixedACmd `cmd:"" name:"a-tag"`
		B *signatureMixedBCmd `cmd:"" help:"b-tag-help"`
	} `embed:"" group:"Parent Group" prefix:"p-"`
}

type signatureSplitSourcesCmd struct{}

func (signatureSplitSourcesCmd) Signature() string {
	return `name:"run" help:"sig-help" aliases:"sig-alias" group:"Child Group"`
}

type signatureSplitSourcesRoot struct {
	Parent struct {
		Cmd *signatureSplitSourcesCmd `cmd:"" aliases:"tag-alias"`
	} `embed:"" group:"Parent Group" prefix:"x-"`
}

type expectedCommand struct {
	name     string
	help     string
	aliases  string
	groupKey string
	hidden   *bool
}

func assertCommand(t *testing.T, label string, cmd *Node, want expectedCommand) {
	t.Helper()
	if got := cmd.Name; got != want.name {
		t.Fatalf("%s name = %q, want %q", label, got, want.name)
	}
	if got := cmd.Help; got != want.help {
		t.Fatalf("%s help = %q, want %q", label, got, want.help)
	}
	if got := strings.Join(cmd.Aliases, ","); got != want.aliases {
		t.Fatalf("%s aliases = %q, want %q", label, got, want.aliases)
	}
	if cmd.Group == nil || cmd.Group.Key != want.groupKey {
		key := ""
		if cmd.Group != nil {
			key = cmd.Group.Key
		}
		t.Fatalf("%s group key = %q, want %q", label, key, want.groupKey)
	}
	if want.hidden != nil && cmd.Hidden != *want.hidden {
		t.Fatalf("%s hidden = %v, want %v", label, cmd.Hidden, *want.hidden)
	}
}

func TestSignatureMethodSingleCommandScenarios(t *testing.T) {
	trueVal := true
	falseVal := false
	tests := []struct {
		name string
		root func() any
		want expectedCommand
	}{
		{
			name: "signature adds hidden while field tags win overlaps",
			root: func() any { return &signatureMethodRoot{Cmd: &signatureMethodCmd{}} },
			want: expectedCommand{
				name: "tag", help: "tag help", aliases: "t1,t2", groupKey: "Tag Group", hidden: &trueVal,
			},
		},
		{
			name: "empty signature keeps field tags",
			root: func() any { return &signatureEmptyRoot{Cmd: &signatureEmptyCmd{}} },
			want: expectedCommand{
				name: "tag", help: "tag help", aliases: "t1,t2", groupKey: "Tag Group", hidden: &falseVal,
			},
		},
		{
			name: "signature retains inherited parent group",
			root: func() any {
				return &signatureInheritedGroupRoot{Nested: struct {
					Cmd *signatureInheritedGroupCmd `cmd:""`
				}{
					Cmd: &signatureInheritedGroupCmd{},
				}}
			},
			want: expectedCommand{
				name: "cmd", help: "from signature", aliases: "", groupKey: "Parent Group",
			},
		},
		{
			name: "parent field and child signature composition",
			root: func() any {
				return &signatureSplitSourcesRoot{Parent: struct {
					Cmd *signatureSplitSourcesCmd `cmd:"" aliases:"tag-alias"`
				}{
					Cmd: &signatureSplitSourcesCmd{},
				}}
			},
			want: expectedCommand{
				name: "x-run", help: "sig-help", aliases: "tag-alias", groupKey: "Child Group",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, err := New(tt.root())
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			if got, want := len(k.Model.Node.Children), 1; got != want {
				t.Fatalf("children = %d, want %d", got, want)
			}
			assertCommand(t, "cmd", k.Model.Node.Children[0], tt.want)
		})
	}
}

func TestSignatureMethodMultipleCommandsWithMixedSources(t *testing.T) {
	root := &signatureMixedRoot{Parent: struct {
		A *signatureMixedACmd `cmd:"" name:"a-tag"`
		B *signatureMixedBCmd `cmd:"" help:"b-tag-help"`
	}{
		A: &signatureMixedACmd{},
		B: &signatureMixedBCmd{},
	}}
	k, err := New(root)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	if got, want := len(k.Model.Node.Children), 2; got != want {
		t.Fatalf("children = %d, want %d", got, want)
	}

	trueVal := true
	wantByName := map[string]expectedCommand{
		"p-a-tag": {name: "p-a-tag", help: "sig-a-help", aliases: "sig-a", groupKey: "Parent Group"},
		"p-sig-b": {name: "p-sig-b", help: "b-tag-help", aliases: "", groupKey: "Parent Group", hidden: &trueVal},
	}
	for _, cmd := range k.Model.Node.Children {
		want, ok := wantByName[cmd.Name]
		if !ok {
			t.Fatalf("unexpected command %q", cmd.Name)
		}
		assertCommand(t, cmd.Name, cmd, want)
		delete(wantByName, cmd.Name)
	}
	if len(wantByName) != 0 {
		t.Fatalf("missing commands: %v", wantByName)
	}
}

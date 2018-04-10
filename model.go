package kong

import "flag"

type Value = flag.Getter

type ApplicationModel struct {
	Name        string
	Description string

	NodeModel
}

type NodeModel struct {
	Groups []*GroupModel
	// Positional arguments.
	Arguments []*ArgumentModel
}

type GroupModel struct {
	// Flags.
	Flags []*FlagModel
	// Command hierarchy.
	Commands []*CommandModel
}

type ValueModel struct {
	Name  string
	Help  string
	Value flag.Value
}

type CommandModel struct {
	NodeModel
}

type ArgumentModel struct {
	ValueModel
}

type FlagModel struct {
	ValueModel
	Short rune
}

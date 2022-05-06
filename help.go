package kong

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"strings"
)

const (
	defaultIndent        = 2
	defaultColumnPadding = 4
)

// Help flag.
type helpValue bool

func (h helpValue) BeforeApply(ctx *Context) error {
	options := ctx.Kong.helpOptions
	options.Summary = false
	err := ctx.Kong.help(options, ctx)
	if err != nil {
		return err
	}
	ctx.Kong.Exit(0)
	return nil
}

// HelpOptions for HelpPrinters.
type HelpOptions struct {
	// Don't print top-level usage summary.
	NoAppSummary bool

	// Write a one-line summary of the context.
	Summary bool

	// Write help in a more compact, but still fully-specified, form.
	Compact bool

	// Tree writes command chains in a tree structure instead of listing them separately.
	Tree bool

	// Place the flags after the commands listing.
	FlagsLast bool

	// Indenter modulates the given prefix for the next layer in the tree view.
	// The following exported templates can be used: kong.SpaceIndenter, kong.LineIndenter, kong.TreeIndenter
	// The kong.SpaceIndenter will be used by default.
	Indenter HelpIndenter

	// Don't show the help associated with subcommands
	NoExpandSubcommands bool

	// Clamp the help wrap width to a value smaller than the terminal width.
	// If this is set to a non-positive number, the terminal width is used; otherwise,
	// the min of this value or the terminal width is used.
	WrapUpperBound int
}

// Apply options to Kong as a configuration option.
func (h HelpOptions) Apply(k *Kong) error {
	k.helpOptions = h
	return nil
}

// HelpProvider can be implemented by commands/args to provide detailed help.
type HelpProvider interface {
	// This string is formatted by go/doc and thus has the same formatting rules.
	Help() string
}

// PlaceHolderProvider can be implemented by mappers to provide custom placeholder text.
type PlaceHolderProvider interface {
	PlaceHolder(flag *Flag) string
}

// HelpIndenter is used to indent new layers in the help tree.
type HelpIndenter func(prefix string) string

// HelpPrinter is used to print context-sensitive help.
type HelpPrinter func(options HelpOptions, ctx *Context) error

// HelpValueFormatter is used to format the help text of flags and positional arguments.
type HelpValueFormatter func(value *Value) string

// DefaultHelpValueFormatter is the default HelpValueFormatter.
func DefaultHelpValueFormatter(value *Value) string {
	if value.Tag.Env == "" || HasInterpolatedVar(value.OrigHelp, "env") {
		return value.Help
	}
	suffix := "($" + value.Tag.Env + ")"
	switch {
	case strings.HasSuffix(value.Help, "."):
		return value.Help[:len(value.Help)-1] + " " + suffix + "."
	case value.Help == "":
		return suffix
	default:
		return value.Help + " " + suffix
	}
}

// DefaultShortHelpPrinter is the default HelpPrinter for short help on error.
func DefaultShortHelpPrinter(options HelpOptions, ctx *Context) error {
	w := NewHelpWriter(ctx, options)
	cmd := ctx.Selected()
	app := ctx.Model
	if cmd == nil {
		w.Printf("Usage: %s%s", app.Name, app.Summary())
		w.Printf(`Run "%s --help" for more information.`, app.Name)
	} else {
		w.Printf("Usage: %s %s", app.Name, cmd.Summary())
		w.Printf(`Run "%s --help" for more information.`, cmd.FullPath())
	}
	return w.Write(ctx.Stdout)
}

// DefaultHelpPrinter is the default HelpPrinter.
func DefaultHelpPrinter(options HelpOptions, ctx *Context) error {
	if ctx.Empty() {
		options.Summary = false
	}
	w := NewHelpWriter(ctx, options)
	selected := ctx.Selected()
	if selected == nil {
		WriteApp(w, ctx.Model)
	} else {
		WriteCommand(w, ctx.Model, selected)
	}
	return w.Write(ctx.Stdout)
}

// WriteApp writes the help output for the whole app.
func WriteApp(w *HelpWriter, app *Application) {
	if !w.NoAppSummary {
		w.Printf("Usage: %s%s", app.Name, app.Summary())
	}
	WriteNodeDetail(w, app.Node, true)
	cmds := app.Leaves(true)
	if len(cmds) > 0 && app.HelpFlag != nil {
		w.Print("")
		if w.Summary {
			w.Printf(`Run "%s --help" for more information.`, app.Name)
		} else {
			w.Printf(`Run "%s <command> --help" for more information on a command.`, app.Name)
		}
	}
}

// WriteCommand writes the help output for a specific command.
func WriteCommand(w *HelpWriter, app *Application, cmd *Command) {
	if !w.NoAppSummary {
		w.Printf("Usage: %s %s", app.Name, cmd.Summary())
	}
	WriteNodeDetail(w, cmd, true)
	if w.Summary && app.HelpFlag != nil {
		w.Print("")
		w.Printf(`Run "%s --help" for more information.`, cmd.FullPath())
	}
}

// WriteNodeDetail writes the help output for a node.
func WriteNodeDetail(w *HelpWriter, node *Node, hide bool) {
	if node.Help != "" {
		w.Print("")
		w.Wrap(node.Help)
	}
	if w.Summary {
		return
	}
	if node.Detail != "" {
		w.Print("")
		w.Wrap(node.Detail)
	}
	if len(node.Positional) > 0 {
		w.Print("")
		w.WritePositionals(node.Positional, "Arguments:")
	}
	if !w.FlagsLast {
		w.WriteFlagsForCommand(node, "Flags:")
	}
	var cmds []*Node
	if w.NoExpandSubcommands {
		cmds = node.Children
	} else {
		cmds = node.Leaves(hide)
	}
	if len(cmds) > 0 {
		if w.Tree {
			w.Print("")
			w.WriteCommandTree(node, "Commands:")
		} else {
			groupedCmds := collectCommandGroups(cmds)
			for _, group := range groupedCmds {
				w.Print("")
				if group.Metadata.Title != "" {
					w.Wrap(group.Metadata.Title)
				}
				if group.Metadata.Description != "" {
					w.Indent()
					w.Wrap(group.Metadata.Description)
					w.Unindent()
					w.Print("")
				}

				if w.Compact {
					w.Indent()
					w.WriteCompactCommandList(group.Commands)
					w.Unindent()
				} else {
					w.Indent()
					w.WriteCommandList(group.Commands)
					w.Unindent()
				}
			}
		}
	}
	if w.FlagsLast {
		w.WriteFlagsForCommand(node, "Flags:")
	}
}

// WriteFlagsForCommand writes the help output for flags.
func (w *HelpWriter) WriteFlagsForCommand(node *Node, title string) {
	if flags := node.AllFlags(true); len(flags) > 0 {
		groupedFlags := collectFlagGroups(title, flags)
		for _, group := range groupedFlags {
			w.Print("")
			if group.Metadata.Title != "" {
				w.Wrap(group.Metadata.Title)
			}
			if group.Metadata.Description != "" {
				w.Indent()
				w.Wrap(group.Metadata.Description)
				w.Unindent()
				w.Print("")
			}
			w.Indent()
			w.WriteFlags(group.Flags)
			w.Unindent()
		}
	}
}

// WriteCommandList writes the help output for the command list.
func (w *HelpWriter) WriteCommandList(cmds []*Node) {
	for i, cmd := range cmds {
		if cmd.Hidden {
			continue
		}
		w.WriteCommandSummary(cmd)
		if i != len(cmds)-1 {
			w.Print("")
		}
	}
}

// WriteCompactCommandList writes the help output for the compact command list.
func (w *HelpWriter) WriteCompactCommandList(cmds []*Node) {
	rows := [][2]string{}
	for _, cmd := range cmds {
		if cmd.Hidden {
			continue
		}
		rows = append(rows, [2]string{cmd.Path(), cmd.Help})
	}
	w.WriteTwoColumns(rows)
}

// WriteCommandTree writes the help output for the command tree.
func (w *HelpWriter) WriteCommandTree(node *Node, title string) {
	if title != "" {
		w.Print(title)
	}
	rows := make([][2]string, 0, len(node.Children)*2)
	for i, cmd := range node.Children {
		if cmd.Hidden {
			continue
		}
		rows = append(rows, w.CommandTree(cmd, "")...)
		if i != len(node.Children)-1 {
			rows = append(rows, [2]string{"", ""})
		}
	}
	w.Indent()
	w.WriteTwoColumns(rows)
	w.Unindent()
}

type helpFlagGroup struct {
	Metadata *Group
	Flags    [][]*Flag
}

func collectFlagGroups(title string, flags [][]*Flag) []helpFlagGroup {
	// Group keys in order of appearance.
	groups := []*Group{}
	// Flags grouped by their group key.
	flagsByGroup := map[string][][]*Flag{}

	for _, levelFlags := range flags {
		levelFlagsByGroup := map[string][]*Flag{}

		for _, flag := range levelFlags {
			key := ""
			if flag.Group != nil {
				key = flag.Group.Key
				groupAlreadySeen := false
				for _, group := range groups {
					if key == group.Key {
						groupAlreadySeen = true
						break
					}
				}
				if !groupAlreadySeen {
					groups = append(groups, flag.Group)
				}
			}

			levelFlagsByGroup[key] = append(levelFlagsByGroup[key], flag)
		}

		for key, flags := range levelFlagsByGroup {
			flagsByGroup[key] = append(flagsByGroup[key], flags)
		}
	}

	out := []helpFlagGroup{}
	// Ungrouped flags are always displayed first.
	if ungroupedFlags, ok := flagsByGroup[""]; ok {
		out = append(out, helpFlagGroup{
			Metadata: &Group{Title: title},
			Flags:    ungroupedFlags,
		})
	}
	for _, group := range groups {
		out = append(out, helpFlagGroup{Metadata: group, Flags: flagsByGroup[group.Key]})
	}
	return out
}

type helpCommandGroup struct {
	Metadata *Group
	Commands []*Node
}

func collectCommandGroups(nodes []*Node) []helpCommandGroup {
	// Groups in order of appearance.
	groups := []*Group{}
	// Nodes grouped by their group key.
	nodesByGroup := map[string][]*Node{}

	for _, node := range nodes {
		key := ""
		if group := node.ClosestGroup(); group != nil {
			key = group.Key
			if _, ok := nodesByGroup[key]; !ok {
				groups = append(groups, group)
			}
		}
		nodesByGroup[key] = append(nodesByGroup[key], node)
	}

	out := []helpCommandGroup{}
	// Ungrouped nodes are always displayed first.
	if ungroupedNodes, ok := nodesByGroup[""]; ok {
		out = append(out, helpCommandGroup{
			Metadata: &Group{Title: "Commands:"},
			Commands: ungroupedNodes,
		})
	}
	for _, group := range groups {
		out = append(out, helpCommandGroup{Metadata: group, Commands: nodesByGroup[group.Key]})
	}
	return out
}

// WriteCommandSummary prints the summary to HelpWriter
func (w *HelpWriter) WriteCommandSummary(cmd *Command) {
	w.Print(cmd.Summary())
	if cmd.Help != "" {
		w.Indent()
		w.Wrap(cmd.Help)
		w.Unindent()
	}
}

// HelpWriter is used to write help output.
type HelpWriter struct {
	indent          string
	IndentCharacter string
	width           int
	lines           *[]string
	helpFormatter   HelpValueFormatter
	HelpOptions
}

// NewHelpWriter creates a new HelpWriter.
func NewHelpWriter(ctx *Context, options HelpOptions) *HelpWriter {
	lines := []string{}
	wrapWidth := guessWidth(ctx.Stdout)
	if options.WrapUpperBound > 0 && wrapWidth > options.WrapUpperBound {
		wrapWidth = options.WrapUpperBound
	}
	w := &HelpWriter{
		indent:          "",
		IndentCharacter: strings.Repeat(" ", defaultIndent),
		width:           wrapWidth,
		lines:           &lines,
		helpFormatter:   ctx.Kong.helpFormatter,
		HelpOptions:     options,
	}
	return w
}

// Printf writes the formated input to a line.
func (w *HelpWriter) Printf(format string, args ...interface{}) {
	w.Print(fmt.Sprintf(format, args...))
}

// Print writes the text to a line.
func (w *HelpWriter) Print(text string) {
	*w.lines = append(*w.lines, strings.TrimRight(w.indent+text, " "))
}

// Indent adds an indentation by IndentCharacter(defaults to two spaces).
func (w *HelpWriter) Indent() {
	w.indent += w.IndentCharacter
	w.width -= len(w.IndentCharacter)
}

// Unindent removes an indentation by IndentCharacter(defaults to two spaces).
func (w *HelpWriter) Unindent() {
	w.indent = w.indent[:len(w.indent)-len(w.IndentCharacter)]
	w.width += len(w.IndentCharacter)
}

func (w *HelpWriter) String() string {
	return strings.Join(*w.lines, "\n")
}

func (w *HelpWriter) Write(writer io.Writer) error {
	for _, line := range *w.lines {
		_, err := io.WriteString(writer, line+"\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// Wrap will write the text after adding new lines based on width of the terminal.
func (w *HelpWriter) Wrap(text string) {
	buff := bytes.NewBuffer(nil)
	doc.ToText(buff, strings.TrimSpace(text), "", "   ", w.width)
	for _, line := range strings.Split(strings.TrimSpace(buff.String()), "\n") {
		w.Print(line)
	}
}

// WritePositionals writes positional to HelpWriter.
func (w *HelpWriter) WritePositionals(args []*Positional, title string) {
	if title != "" {
		w.Print(title)
	}
	rows := [][2]string{}
	for _, arg := range args {
		rows = append(rows, [2]string{arg.Summary(), w.helpFormatter(arg)})
	}
	w.Indent()
	w.WriteTwoColumns(rows)
	w.Unindent()
}

// WriteFlags writes the flags to HelpWriter.
func (w *HelpWriter) WriteFlags(groups [][]*Flag) {
	rows := [][2]string{}
	haveShort := false
	for _, group := range groups {
		for _, flag := range group {
			if flag.Short != 0 {
				haveShort = true
				break
			}
		}
	}
	for i, group := range groups {
		if i > 0 {
			rows = append(rows, [2]string{"", ""})
		}
		for _, flag := range group {
			if !flag.Hidden {
				rows = append(rows, [2]string{FormatFlag(haveShort, flag), w.helpFormatter(flag.Value)})
			}
		}
	}
	w.WriteTwoColumns(rows)
}

// WriteTwoColumns writes two columns.
func (w *HelpWriter) WriteTwoColumns(rows [][2]string) {
	maxLeft := 375 * w.width / 1000
	if maxLeft < 30 {
		maxLeft = 30
	}
	// Find size of first column.
	leftSize := 0
	for _, row := range rows {
		if c := len(row[0]); c > leftSize && c < maxLeft {
			leftSize = c
		}
	}

	offsetStr := strings.Repeat(" ", leftSize+defaultColumnPadding)

	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", strings.Repeat(" ", defaultIndent), w.width-leftSize-defaultColumnPadding)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")

		line := fmt.Sprintf("%-*s", leftSize, row[0])
		if len(row[0]) < maxLeft {
			line += fmt.Sprintf("%*s%s", defaultColumnPadding, "", lines[0])
			lines = lines[1:]
		}
		w.Print(line)
		for _, line := range lines {
			w.Printf("%s%s", offsetStr, line)
		}
	}
}

// FormatFlag formats the flag for the help output.
// haveShort will be true if there are short flags present at all in the help. Useful for column alignment.
func FormatFlag(haveShort bool, flag *Flag) string {
	flagString := ""
	name := flag.Name
	isBool := flag.IsBool()
	if flag.Short != 0 {
		if isBool && flag.Tag.Negatable {
			flagString += fmt.Sprintf("-%c, --[no-]%s", flag.Short, name)
		} else {
			flagString += fmt.Sprintf("-%c, --%s", flag.Short, name)
		}
	} else {
		if isBool && flag.Tag.Negatable {
			if haveShort {
				flagString = fmt.Sprintf("    --[no-]%s", name)
			} else {
				flagString = fmt.Sprintf("--[no-]%s", name)
			}
		} else {
			if haveShort {
				flagString += fmt.Sprintf("    --%s", name)
			} else {
				flagString += fmt.Sprintf("--%s", name)
			}
		}
	}
	if !isBool {
		flagString += fmt.Sprintf("=%s", flag.FormatPlaceHolder())
	}
	return flagString
}

// CommandTree creates a tree with the given node name as root and its children's arguments and sub commands as leaves.
func (h *HelpOptions) CommandTree(node *Node, prefix string) (rows [][2]string) {
	var nodeName string
	switch node.Type {
	default:
		nodeName += prefix + node.Name
		if len(node.Aliases) != 0 {
			nodeName += fmt.Sprintf(" (%s)", strings.Join(node.Aliases, ","))
		}
	case ArgumentNode:
		nodeName += prefix + "<" + node.Name + ">"
	}
	rows = append(rows, [2]string{nodeName, node.Help})
	if h.Indenter == nil {
		prefix = SpaceIndenter(prefix)
	} else {
		prefix = h.Indenter(prefix)
	}
	for _, arg := range node.Positional {
		rows = append(rows, [2]string{prefix + arg.Summary(), arg.Help})
	}
	for _, subCmd := range node.Children {
		if subCmd.Hidden {
			continue
		}
		rows = append(rows, h.CommandTree(subCmd, prefix)...)
	}
	return
}

// SpaceIndenter adds a space indent to the given prefix.
func SpaceIndenter(prefix string) string {
	return prefix + strings.Repeat(" ", defaultIndent)
}

// LineIndenter adds line points to every new indent.
func LineIndenter(prefix string) string {
	if prefix == "" {
		return "- "
	}
	return strings.Repeat(" ", defaultIndent) + prefix
}

// TreeIndenter adds line points to every new indent and vertical lines to every layer.
func TreeIndenter(prefix string) string {
	if prefix == "" {
		return "|- "
	}
	return "|" + strings.Repeat(" ", defaultIndent) + prefix
}

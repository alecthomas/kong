// nolint
package main

import "fmt"

type AttachCmd struct {
	DetachKeys string `help:"Override the key sequence for detaching a container"`
	NoStdin    bool   `help:"Do not attach STDIN"`
	SigProxy   bool   `help:"Proxy all received signals to the process" default:"true"`

	Container string `arg required help:"Container ID to attach to."`
}

func (a *AttachCmd) Run(globals *Globals) error {
	fmt.Printf("Config: %s\n", globals.Config)
	fmt.Printf("Attaching to: %v\n", a.Container)
	fmt.Printf("SigProxy: %v\n", a.SigProxy)
	return nil
}

type BuildCmd struct {
	Arg string `arg required`
}

func (cmd *BuildCmd) Run(globals *Globals) error {
	return nil
}

type CommitCmd struct {
	Arg string `arg required`
}

func (cmd *CommitCmd) Run(globals *Globals) error {
	return nil
}

type CpCmd struct {
	Arg string `arg required`
}

func (cmd *CpCmd) Run(globals *Globals) error {
	return nil
}

type CreateCmd struct {
	Arg string `arg required`
}

func (cmd *CreateCmd) Run(globals *Globals) error {
	return nil
}

type DeployCmd struct {
	Arg string `arg required`
}

func (cmd *DeployCmd) Run(globals *Globals) error {
	return nil
}

type DiffCmd struct {
	Arg string `arg required`
}

func (cmd *DiffCmd) Run(globals *Globals) error {
	return nil
}

type EventsCmd struct {
	Arg string `arg required`
}

func (cmd *EventsCmd) Run(globals *Globals) error {
	return nil
}

type ExecCmd struct {
	Arg string `arg required`
}

func (cmd *ExecCmd) Run(globals *Globals) error {
	return nil
}

type ExportCmd struct {
	Arg string `arg required`
}

func (cmd *ExportCmd) Run(globals *Globals) error {
	return nil
}

type HistoryCmd struct {
	Arg string `arg required`
}

func (cmd *HistoryCmd) Run(globals *Globals) error {
	return nil
}

type ImagesCmd struct {
	Arg string `arg required`
}

func (cmd *ImagesCmd) Run(globals *Globals) error {
	return nil
}

type ImportCmd struct {
	Arg string `arg required`
}

func (cmd *ImportCmd) Run(globals *Globals) error {
	return nil
}

type InfoCmd struct {
	Arg string `arg required`
}

func (cmd *InfoCmd) Run(globals *Globals) error {
	return nil
}

type InspectCmd struct {
	Arg string `arg required`
}

func (cmd *InspectCmd) Run(globals *Globals) error {
	return nil
}

type KillCmd struct {
	Arg string `arg required`
}

func (cmd *KillCmd) Run(globals *Globals) error {
	return nil
}

type LoadCmd struct {
	Arg string `arg required`
}

func (cmd *LoadCmd) Run(globals *Globals) error {
	return nil
}

type LoginCmd struct {
	Arg string `arg required`
}

func (cmd *LoginCmd) Run(globals *Globals) error {
	return nil
}

type LogoutCmd struct {
	Arg string `arg required`
}

func (cmd *LogoutCmd) Run(globals *Globals) error {
	return nil
}

type LogsCmd struct {
	Arg string `arg required`
}

func (cmd *LogsCmd) Run(globals *Globals) error {
	return nil
}

type PauseCmd struct {
	Arg string `arg required`
}

func (cmd *PauseCmd) Run(globals *Globals) error {
	return nil
}

type PortCmd struct {
	Arg string `arg required`
}

func (cmd *PortCmd) Run(globals *Globals) error {
	return nil
}

type PsCmd struct {
	Arg string `arg required`
}

func (cmd *PsCmd) Run(globals *Globals) error {
	return nil
}

type PullCmd struct {
	Arg string `arg required`
}

func (cmd *PullCmd) Run(globals *Globals) error {
	return nil
}

type PushCmd struct {
	Arg string `arg required`
}

func (cmd *PushCmd) Run(globals *Globals) error {
	return nil
}

type RenameCmd struct {
	Arg string `arg required`
}

func (cmd *RenameCmd) Run(globals *Globals) error {
	return nil
}

type RestartCmd struct {
	Arg string `arg required`
}

func (cmd *RestartCmd) Run(globals *Globals) error {
	return nil
}

type RmCmd struct {
	Arg string `arg required`
}

func (cmd *RmCmd) Run(globals *Globals) error {
	return nil
}

type RmiCmd struct {
	Arg string `arg required`
}

func (cmd *RmiCmd) Run(globals *Globals) error {
	return nil
}

type RunCmd struct {
	Arg string `arg required`
}

func (cmd *RunCmd) Run(globals *Globals) error {
	return nil
}

type SaveCmd struct {
	Arg string `arg required`
}

func (cmd *SaveCmd) Run(globals *Globals) error {
	return nil
}

type SearchCmd struct {
	Arg string `arg required`
}

func (cmd *SearchCmd) Run(globals *Globals) error {
	return nil
}

type StartCmd struct {
	Arg string `arg required`
}

func (cmd *StartCmd) Run(globals *Globals) error {
	return nil
}

type StatsCmd struct {
	Arg string `arg required`
}

func (cmd *StatsCmd) Run(globals *Globals) error {
	return nil
}

type StopCmd struct {
	Arg string `arg required`
}

func (cmd *StopCmd) Run(globals *Globals) error {
	return nil
}

type TagCmd struct {
	Arg string `arg required`
}

func (cmd *TagCmd) Run(globals *Globals) error {
	return nil
}

type TopCmd struct {
	Arg string `arg required`
}

func (cmd *TopCmd) Run(globals *Globals) error {
	return nil
}

type UnpauseCmd struct {
	Arg string `arg required`
}

func (cmd *UnpauseCmd) Run(globals *Globals) error {
	return nil
}

type UpdateCmd struct {
	Arg string `arg required`
}

func (cmd *UpdateCmd) Run(globals *Globals) error {
	return nil
}

type VersionCmd struct {
	Arg string `arg required`
}

func (cmd *VersionCmd) Run(globals *Globals) error {
	return nil
}

type WaitCmd struct {
	Arg string `arg required`
}

func (cmd *WaitCmd) Run(globals *Globals) error {
	return nil
}

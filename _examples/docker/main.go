// nolint
package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type Globals struct {
	Config    string      `help:"Location of client config files" default:"~/.docker" type:"path"`
	Debug     bool        `short:"D" help:"Enable debug mode"`
	Host      []string    `short:"H" help:"Daemon socket(s) to connect to"`
	LogLevel  string      `short:"l" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
	TLS       bool        `help:"Use TLS; implied by --tls-verify"`
	TLSCACert string      `name:"tls-ca-cert" help:"Trust certs signed only by this CA" default:"~/.docker/ca.pem" type:"path"`
	TLSCert   string      `help:"Path to TLS certificate file" default:"~/.docker/cert.pem" type:"path"`
	TLSKey    string      `help:"Path to TLS key file" default:"~/.docker/key.pem" type:"path"`
	TLSVerify bool        `help:"Use TLS and verify the remote"`
	Version   VersionFlag `name:"version" help:"Print version information and quit"`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	Globals

	Attach  AttachCmd  `cmd:"" help:"Attach local standard input, output, and error streams to a running container"`
	Build   BuildCmd   `cmd:"" help:"Build an image from a Dockerfile"`
	Commit  CommitCmd  `cmd:"" help:"Create a new image from a container's changes"`
	Cp      CpCmd      `cmd:"" help:"Copy files/folders between a container and the local filesystem"`
	Create  CreateCmd  `cmd:"" help:"Create a new container"`
	Deploy  DeployCmd  `cmd:"" help:"Deploy a new stack or update an existing stack"`
	Diff    DiffCmd    `cmd:"" help:"Inspect changes to files or directories on a container's filesystem"`
	Events  EventsCmd  `cmd:"" help:"Get real time events from the server"`
	Exec    ExecCmd    `cmd:"" help:"Run a command in a running container"`
	Export  ExportCmd  `cmd:"" help:"Export a container's filesystem as a tar archive"`
	History HistoryCmd `cmd:"" help:"Show the history of an image"`
	Images  ImagesCmd  `cmd:"" help:"List images"`
	Import  ImportCmd  `cmd:"" help:"Import the contents from a tarball to create a filesystem image"`
	Info    InfoCmd    `cmd:"" help:"Display system-wide information"`
	Inspect InspectCmd `cmd:"" help:"Return low-level information on Docker objects"`
	Kill    KillCmd    `cmd:"" help:"Kill one or more running containers"`
	Load    LoadCmd    `cmd:"" help:"Load an image from a tar archive or STDIN"`
	Login   LoginCmd   `cmd:"" help:"Log in to a Docker registry"`
	Logout  LogoutCmd  `cmd:"" help:"Log out from a Docker registry"`
	Logs    LogsCmd    `cmd:"" help:"Fetch the logs of a container"`
	Pause   PauseCmd   `cmd:"" help:"Pause all processes within one or more containers"`
	Port    PortCmd    `cmd:"" help:"List port mappings or a specific mapping for the container"`
	Ps      PsCmd      `cmd:"" help:"List containers"`
	Pull    PullCmd    `cmd:"" help:"Pull an image or a repository from a registry"`
	Push    PushCmd    `cmd:"" help:"Push an image or a repository to a registry"`
	Rename  RenameCmd  `cmd:"" help:"Rename a container"`
	Restart RestartCmd `cmd:"" help:"Restart one or more containers"`
	Rm      RmCmd      `cmd:"" help:"Remove one or more containers"`
	Rmi     RmiCmd     `cmd:"" help:"Remove one or more images"`
	Run     RunCmd     `cmd:"" help:"Run a command in a new container"`
	Save    SaveCmd    `cmd:"" help:"Save one or more images to a tar archive (streamed to STDOUT by default)"`
	Search  SearchCmd  `cmd:"" help:"Search the Docker Hub for images"`
	Start   StartCmd   `cmd:"" help:"Start one or more stopped containers"`
	Stats   StatsCmd   `cmd:"" help:"Display a live stream of container(s) resource usage statistics"`
	Stop    StopCmd    `cmd:"" help:"Stop one or more running containers"`
	Tag     TagCmd     `cmd:"" help:"Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE"`
	Top     TopCmd     `cmd:"" help:"Display the running processes of a container"`
	Unpause UnpauseCmd `cmd:"" help:"Unpause all processes within one or more containers"`
	Update  UpdateCmd  `cmd:"" help:"Update configuration of one or more containers"`
	Version VersionCmd `cmd:"" help:"Show the Docker version information"`
	Wait    WaitCmd    `cmd:"" help:"Block until one or more containers stop, then print their exit codes"`
}

func main() {
	cli := CLI{
		Globals: Globals{
			Version: VersionFlag("0.1.1"),
		},
	}

	ctx := kong.Parse(&cli,
		kong.Name("docker"),
		kong.Description("A self-sufficient runtime for containers"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		})
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}

// nolint
package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

type AttachCmd struct {
	DetachKeys string `help:"Override the key sequence for detaching a container"`
	NoStdin    bool   `help:"Do not attach STDIN"`
	SigProxy   bool   `help:"Proxy all received signals to the process" default:"true"`

	Container string `arg required help:"Container ID to attach to."`
}

func (a *AttachCmd) Run() error {
	fmt.Printf("Attaching to: %v\n", a.Container)
	fmt.Printf("SigProxy: %v\n", a.SigProxy)
	return nil
}

var cli struct {
	Config       string   `help:"Location of client config files" default:"~/.docker" type:"path"`
	Debug        bool     `short:"D" help:"Enable debug mode"`
	Host         []string `short:"H" help:"Daemon socket(s) to connect to"`
	LogLevel     string   `short:"l" help:"Set the logging level (debug|info|warn|error|fatal)" default:"info"`
	TLS          bool     `help:"Use TLS; implied by --tlsverify"`
	TLSCACert    string   `name:"tls-ca-cert" help:"Trust certs signed only by this CA" default:"~/.docker/ca.pem" type:"path"`
	TLSCert      string   `name:"tls-cert" help:"Path to TLS certificate file" default:"~/.docker/cert.pem" type:"path"`
	TLSKey       string   `help:"Path to TLS key file" default:"~/.docker/key.pem" type:"path"`
	TLSVerify    bool     `help:"Use TLS and verify the remote"`
	PrintVersion bool     `name:"version" help:"Print version information and quit"`

	Attach AttachCmd `cmd help:"Attach local standard input, output, and error streams to a running container"`

	Build struct {
		Arg string `arg required`
	} `cmd help:"Build an image from a Dockerfile"`
	Commit struct {
		Arg string `arg required`
	} `cmd help:"Create a new image from a container's changes"`
	Cp struct {
		Arg string `arg required`
	} `cmd help:"Copy files/folders between a container and the local filesystem"`
	Create struct {
		Arg string `arg required`
	} `cmd help:"Create a new container"`
	Deploy struct {
		Arg string `arg required`
	} `cmd help:"Deploy a new stack or update an existing stack"`
	Diff struct {
		Arg string `arg required`
	} `cmd help:"Inspect changes to files or directories on a container's filesystem"`
	Events struct {
		Arg string `arg required`
	} `cmd help:"Get real time events from the server"`
	Exec struct {
		Arg string `arg required`
	} `cmd help:"Run a command in a running container"`
	Export struct {
		Arg string `arg required`
	} `cmd help:"Export a container's filesystem as a tar archive"`
	History struct {
		Arg string `arg required`
	} `cmd help:"Show the history of an image"`
	Images struct {
		Arg string `arg required`
	} `cmd help:"List images"`
	Import struct {
		Arg string `arg required`
	} `cmd help:"Import the contents from a tarball to create a filesystem image"`
	Info struct {
		Arg string `arg required`
	} `cmd help:"Display system-wide information"`
	Inspect struct {
		Arg string `arg required`
	} `cmd help:"Return low-level information on Docker objects"`
	Kill struct {
		Arg string `arg required`
	} `cmd help:"Kill one or more running containers"`
	Load struct {
		Arg string `arg required`
	} `cmd help:"Load an image from a tar archive or STDIN"`
	Login struct {
		Arg string `arg required`
	} `cmd help:"Log in to a Docker registry"`
	Logout struct {
		Arg string `arg required`
	} `cmd help:"Log out from a Docker registry"`
	Logs struct {
		Arg string `arg required`
	} `cmd help:"Fetch the logs of a container"`
	Pause struct {
		Arg string `arg required`
	} `cmd help:"Pause all processes within one or more containers"`
	Port struct {
		Arg string `arg required`
	} `cmd help:"List port mappings or a specific mapping for the container"`
	Ps struct {
		Arg string `arg required`
	} `cmd help:"List containers"`
	Pull struct {
		Arg string `arg required`
	} `cmd help:"Pull an image or a repository from a registry"`
	Push struct {
		Arg string `arg required`
	} `cmd help:"Push an image or a repository to a registry"`
	Rename struct {
		Arg string `arg required`
	} `cmd help:"Rename a container"`
	Restart struct {
		Arg string `arg required`
	} `cmd help:"Restart one or more containers"`
	Rm struct {
		Arg string `arg required`
	} `cmd help:"Remove one or more containers"`
	Rmi struct {
		Arg string `arg required`
	} `cmd help:"Remove one or more images"`
	Run struct {
		Arg string `arg required`
	} `cmd help:"Run a command in a new container"`
	Save struct {
		Arg string `arg required`
	} `cmd help:"Save one or more images to a tar archive (streamed to STDOUT by default)"`
	Search struct {
		Arg string `arg required`
	} `cmd help:"Search the Docker Hub for images"`
	Start struct {
		Arg string `arg required`
	} `cmd help:"Start one or more stopped containers"`
	Stats struct {
		Arg string `arg required`
	} `cmd help:"Display a live stream of container(s) resource usage statistics"`
	Stop struct {
		Arg string `arg required`
	} `cmd help:"Stop one or more running containers"`
	Tag struct {
		Arg string `arg required`
	} `cmd help:"Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE"`
	Top struct {
		Arg string `arg required`
	} `cmd help:"Display the running processes of a container"`
	Unpause struct {
		Arg string `arg required`
	} `cmd help:"Unpause all processes within one or more containers"`
	Update struct {
		Arg string `arg required`
	} `cmd help:"Update configuration of one or more containers"`
	Version struct {
		Arg string `arg required`
	} `cmd help:"Show the Docker version information"`
	Wait struct {
		Arg string `arg required`
	} `cmd help:"Block until one or more containers stop, then print their exit codes"`
}

func main() {
	cmd := kong.Parse(&cli,
		kong.Name("docker"),
		kong.Description("A self-sufficient runtime for containers"),
		kong.UsageOnError(),
		kong.HelpOptions(kong.HelpPrinterOptions{
			Compact: true,
		}))
	var err error
	switch cmd {
	case "attach <container>":
		fmt.Println(cli.Config)
		err = cli.Attach.Run()

	default:
		panic("unsupported command " + cmd)
	}
	kong.FatalIfErrorf(err)
}

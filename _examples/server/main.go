package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/gliderlabs/ssh"
	"github.com/google/shlex"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/alecthomas/colour"
	"github.com/alecthomas/kong"
)

// Handle a single SSH interactive connection.
func handle(log *log.Logger, s ssh.Session) error {
	log.Printf("New SSH")
	sshPty, _, isPty := s.Pty()
	if !isPty {
		return errors.New("no PTY requested")
	}
	log.Printf("Using TERM=%s width=%d height=%d", sshPty.Term, sshPty.Window.Width, sshPty.Window.Height)
	cpty, tty, err := pty.Open()
	if err != nil {
		return err
	}
	defer tty.Close()
	state, err := terminal.GetState(int(cpty.Fd()))
	if err != nil {
		return err
	}
	defer terminal.Restore(int(cpty.Fd()), state)

	colour.Fprintln(tty, "^BWelcome!^R")
	go io.Copy(cpty, s)
	go io.Copy(s, cpty)

	parser, err := buildShellParser(tty)
	if err != nil {
		return err
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:             "> ",
		Stderr:             tty,
		Stdout:             tty,
		Stdin:              tty,
		FuncOnWidthChanged: func(f func()) {},
		FuncMakeRaw: func() error {
			_, err := terminal.MakeRaw(int(cpty.Fd())) // nolint: govet
			return err
		},
		FuncExitRaw: func() error { return nil },
	})
	if err != nil {
		return err
	}

	log.Printf("Loop")
	for {
		tty.Sync()

		var line string
		line, err = rl.Readline()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var args []string
		args, err := shlex.Split(string(line))
		if err != nil {
			parser.Errorf("%s", err)
			continue
		}
		var ctx *kong.Context
		ctx, err = parser.Parse(args)
		if err != nil {
			parser.Errorf("%s", err)
			if err, ok := err.(*kong.ParseError); ok {
				log.Println(err.Error())
				err.Context.PrintUsage(false)
			}
			continue
		}
		err = ctx.Run(ctx)
		if err != nil {
			parser.Errorf("%s", err)
			continue
		}
	}
}

func buildShellParser(tty *os.File) (*kong.Kong, error) {
	parser, err := kong.New(&grammar{},
		kong.Name(""),
		kong.Description("Example using Kong for interactive command parsing."),
		kong.Writers(tty, tty),
		kong.Exit(func(int) {}),
		kong.ConfigureHelp(kong.HelpOptions{
			NoAppSummary: true,
		}),
		kong.NoDefaultHelp(),
	)
	return parser, err
}

func handlerWithError(handle func(log *log.Logger, s ssh.Session) error) ssh.Handler {
	return func(s ssh.Session) {
		prefix := fmt.Sprintf("%s->%s ", s.LocalAddr(), s.RemoteAddr())
		l := log.New(os.Stdout, prefix, log.LstdFlags)
		err := handle(l, s)
		if err != nil {
			log.Printf("error: %s", err)
			s.Exit(1)
		} else {
			log.Printf("Bye")
			s.Exit(0)
		}
	}
}

var cli struct {
	HostKey string `type:"existingfile" help:"SSH host key to use." default:"server_rsa_key"`
	Bind    string `help:"Bind address for server." default:"127.0.0.1:6740"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("server"),
		kong.Description("A network server using Kong for interacting with clients."))

	ssh.Handle(handlerWithError(handle))
	log.Printf("SSH listening on: %s", cli.Bind)
	log.Printf("Using host key: %s", cli.HostKey)
	log.Println()
	parts := strings.Split(cli.Bind, ":")
	log.Printf("Connect with: ssh -p %s %s", parts[1], parts[0])
	log.Println()
	err := ssh.ListenAndServe(cli.Bind, nil, ssh.HostKeyFile(cli.HostKey))
	ctx.FatalIfErrorf(err)
}

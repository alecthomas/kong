package kong

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/riywo/loginshell"
)

var shellInstall = map[string]string{
	"bash": "complete -C ${bin} ${cmd}\n",
	"zsh": `autoload -U +X bashcompinit && bashcompinit
complete -C ${bin} ${cmd}
`,
	"fish": `function __complete_${cmd}
    set -lx COMP_LINE (commandline -cp)
    test -z (commandline -ct)
    and set COMP_LINE "$COMP_LINE "
    ${bin}
end
complete -f -c ${cmd} -a "(__complete_${cmd})"
`,
}

// InstallCompletionFlag will install completion to your shell
type InstallCompletionFlag bool

// BeforeApply uninstalls completion into the users shell.
func (c *InstallCompletionFlag) BeforeApply(ctx *Context) error {
	err := InstallCompletionFromContext(ctx)
	if err != nil {
		return err
	}
	ctx.Exit(0)
	return nil
}

// InstallCompletionFromContext writes shell completion for the given command.
func InstallCompletionFromContext(ctx *Context) error {
	shell, err := loginshell.Shell()
	if err != nil {
		return errors.Wrapf(err, "couldn't determine user's shell")
	}
	bin, err := getBinaryPath()
	if err != nil {
		return errors.Wrapf(err, "couldn't find absolute path to ourselves")
	}
	w := ctx.Stdout
	cmd := ctx.Model.Name
	return InstallCompletion(w, shell, cmd, bin)
}

// InstallCompletion writes shell completion for a command.
func InstallCompletion(w io.Writer, shell, cmd, bin string) error {
	script, ok := shellInstall[filepath.Base(shell)]
	if !ok {
		return errors.Errorf("unsupported shell %s", shell)
	}
	vars := map[string]string{"cmd": cmd, "bin": bin}
	fragment := os.Expand(script, func(s string) string {
		v, ok := vars[s]
		if !ok {
			return "$" + s
		}
		return v
	})
	_, err := fmt.Fprint(w, fragment)
	return err
}

func getBinaryPath() (string, error) {
	bin, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(bin)
}

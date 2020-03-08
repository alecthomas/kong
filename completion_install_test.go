package kong

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// nolint: scopelint
func TestInstallCompletion(t *testing.T) {
	tests := map[string]string{
		"zsh":  "autoload -U +X bashcompinit && bashcompinit\ncomplete -C /usr/bin/docker docker\n",
		"bash": "complete -C /usr/bin/docker docker\n",
		"fish": `function __complete_docker
    set -lx COMP_LINE (commandline -cp)
    test -z (commandline -ct)
    and set COMP_LINE "$COMP_LINE "
    /usr/bin/docker
end
complete -f -c docker -a "(__complete_docker)"
`,
	}
	for shell, fragment := range tests {
		t.Run(shell, func(t *testing.T) {
			w := &strings.Builder{}
			err := InstallCompletion(w, shell, "docker", "/usr/bin/docker")
			require.NoError(t, err)
			require.Equal(t, fragment, w.String())
		})
	}
}

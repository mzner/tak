package shell

import "fmt"

const posixHook = `tak() {
    if [ "$1" = "cd" ]; then
        shift
        local dir
        dir=$(command tak cd "$@")
        if [ $? -eq 0 ]; then
            cd "$dir"
        else
            echo "$dir" >&2
            return 1
        fi
    else
        command tak "$@"
    fi
}
`

const fishHook = `function tak
    if test "$argv[1]" = "cd"
        set -l dir (command tak cd $argv[2..])
        if test $status -eq 0
            cd $dir
        else
            echo $dir >&2
            return 1
        end
    else
        command tak $argv
    end
end
`

// Generate returns the shell hook function source for the given shell.
// Supported values: "zsh", "bash", "fish".
func Generate(shellName string) (string, error) {
	switch shellName {
	case "zsh", "bash":
		return posixHook, nil
	case "fish":
		return fishHook, nil
	default:
		return "", fmt.Errorf("unsupported shell: %q (supported: zsh, bash, fish)", shellName)
	}
}

package handler

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

// ConsoleCtxKey is a context key for commands invoked in an interactive console.
var ConsoleCtxKey = "console command"

// Console is a handler for the "console" command.
func Console(cCtx *cli.Context) error {
	reader := bufio.NewReader(os.Stdin)
	file, _ := os.Executable()

	for {
		fmt.Print("kitsh> ")
		text, _ := reader.ReadString('\n')
		args := splitBySpace(strings.TrimSuffix(text, "\n"))
		for len(args) > 0 && strings.HasPrefix(args[0], "-") {
			args = args[1:] // remove any global arguments
		}
		if len(args) == 0 {
			continue
		}

		newArgs := append([]string{file, "--target", cCtx.String("target")}, args...)
		if err := cCtx.App.RunContext(context.WithValue(context.Background(), ConsoleCtxKey, args), newArgs); err != nil {
			color.Red("%s", err)
		}
	}
}

// splitBySpace splits the supplied string on spaces, honoring double-quoted substrings.
func splitBySpace(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})
}

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
		for i := range args {
			// remove leading and trailing quotes
			args[i] = strings.TrimFunc(args[i], func(r rune) bool {
				return r == '\'' || r == '"'
			})
		}
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

// splitBySpace splits the supplied string on spaces, honoring double and single-quoted substrings.
func splitBySpace(s string) []string {
	currentQuoteChar := '\000'
	return strings.FieldsFunc(s, func(r rune) bool {
		if r == '"' || r == '\'' {
			if currentQuoteChar == '\000' {
				currentQuoteChar = r
			} else if currentQuoteChar == r {
				currentQuoteChar = '\000'
			}
		}
		return currentQuoteChar == '\000' && r == ' '
	})
}

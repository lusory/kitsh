package handler

import (
	"context"
	"errors"
	"github.com/peterh/liner"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"strings"
)

// RecursiveConsoleError is an error which gets raised when a user attempts to invoke the console command in a console.
var RecursiveConsoleError = errors.New("the console command cannot be invoked inside of itself")

// ConsoleCtxKey is a context key for commands invoked in an interactive console.
var ConsoleCtxKey = "console command"

// historyFile is the path to the command history file of the interactive console.
var historyFile string

func init() {
	// save the history file into the home dir or the current working directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir, _ = os.Executable()
	}

	historyFile = filepath.Join(homeDir, ".kitsh_history")
}

// Console is a handler for the "console" command.
func Console(cCtx *cli.Context) error {
	if cCtx.Context.Value(ConsoleCtxKey) != nil {
		return RecursiveConsoleError
	}

	file, _ := os.Executable()

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)
	// only completes the base command names for the time being
	line.SetCompleter(func(line string) (c []string) {
		for _, n := range cCtx.App.VisibleCommands() {
			if strings.HasPrefix(n.Name, strings.ToLower(line)) {
				c = append(c, n.Name)
			}
		}
		return
	})

	if !cCtx.Bool("no-history") {
		if f, err := os.Open(historyFile); err == nil {
			line.ReadHistory(f)
			f.Close()
		}

		defer func() {
			if f, err := os.Create(historyFile); err == nil {
				line.WriteHistory(f)
				f.Close()
			} else {
				PrintError("error writing history file: %s\n", err)
			}
		}()
	}

	for {
		if text, err := line.Prompt("kitsh> "); err == nil {
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
			if err := cCtx.App.RunContext(context.WithValue(cCtx.Context, ConsoleCtxKey, args), newArgs); err != nil {
				PrintError("%s\n", err)
			}

			line.AppendHistory(text)
		} else if err == liner.ErrPromptAborted { // Ctrl+C
			break
		} else {
			PrintError("failed to read input: %s\n", err)
		}
	}

	return nil
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

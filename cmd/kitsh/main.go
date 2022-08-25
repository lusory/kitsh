package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/lusory/kitsh/interpreter"
	"os"
)

// main is the application entrypoint.
func main() {
	var target string
	var ssl bool
	var command string
	var interactive bool
	flag.StringVar(&target, "target", "", "the kitsune gRPC target")
	flag.BoolVar(&ssl, "ssl", false, "should an HTTPS connection be established? (default false)")
	flag.StringVar(&command, "command", "", "the command to be interpreted (optional, used for one-liners without the console)")
	flag.BoolVar(&interactive, "interactive", false, "should an interactive console be opened? (default false)")
	flag.Parse()

	if command == "" && !interactive { // would be a no-op, so print help
		flag.Usage()
		return
	}

	intr, err := interpreter.NewInterpreter(target, ssl, true)
	if err != nil {
		color.Red("Connection error: %s", err.Error())
		return
	}
	defer intr.Client.Close()

	if command != "" {
		if out := intr.Interpret(command); out != "" {
			fmt.Println(out)
		}
	}
	if interactive {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("kitsh> ")
			text, _ := reader.ReadString('\n')
			if out := intr.Interpret(text); out != "" {
				fmt.Println(out)
			}
		}
	}
}

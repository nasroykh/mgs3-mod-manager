package main

import (
	"mgs3mod/internal/cli"
	"mgs3mod/internal/manager"
	"os"
)

func main() {
	interactive := cli.InteractiveInput(os.Stdin)
	os.Exit(cli.RunWithInput(os.Args[1:], os.Stdout, os.Stderr, manager.Production(), os.Stdin, interactive))
}

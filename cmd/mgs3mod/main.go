package main

import (
	"mgs3mod/internal/cli"
	"mgs3mod/internal/manager"
	"os"
)

func main() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr, manager.Production())) }

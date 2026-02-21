package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Galdoba/gslog/internal/infrastructure/config"
	"github.com/Galdoba/gslog/internal/presentation/cli/commands"
	"github.com/Galdoba/gslog/internal/presentation/cli/flags"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg := config.Initialize()
	cmd := cli.Command{
		Name:        "gslog",
		Aliases:     []string{},
		Usage:       "manage gslogs",
		Version:     "prototype",
		Description: "log reader utility for gslog package",
		Flags: []cli.Flag{
			flags.GlobalStrict,
		},
		Commands: []*cli.Command{
			commands.ReadLogs(cfg),
		},
		Authors:   []any{"galdoba"},
		Copyright: "",
	}
	if err := cmd.Run(context.TODO(), os.Args); err != nil {
		fmt.Println("program error:", err)
	}
}

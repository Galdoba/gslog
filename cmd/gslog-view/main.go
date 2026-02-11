package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Galdoba/gslog/internal/infrastructure/config"
	"github.com/Galdoba/gslog/internal/presentation/cli/actions"
	"github.com/Galdoba/gslog/internal/presentation/cli/flags"
	"github.com/urfave/cli/v3"
)

func main() {
	cfg := config.Initialize()
	cmd := cli.Command{
		Name:           "gslog-view",
		Aliases:        []string{},
		Usage:          "manage gslogs",
		UsageText:      "",
		ArgsUsage:      "",
		Version:        "",
		Description:    "logmanager for gslog package",
		DefaultCommand: "",
		Category:       "",
		Commands:       []*cli.Command{},
		Flags: []cli.Flag{
			flags.DirFlag,
		},
		HideHelp:                        false,
		HideHelpCommand:                 false,
		HideVersion:                     false,
		EnableShellCompletion:           false,
		ShellCompletionCommandName:      "",
		ShellComplete:                   nil,
		ConfigureShellCompletionCommand: nil,
		Before:                          nil,
		After:                           nil,
		Action:                          actions.View(cfg),
		CommandNotFound:                 nil,
		OnUsageError:                    nil,
		InvalidFlagAccessHandler:        nil,
		Hidden:                          false,
		Authors:                         []any{"galdoba"},
		Copyright:                       "",
		Reader:                          os.Stdin,
		Writer:                          nil,
		ErrWriter:                       nil,
		ExitErrHandler:                  nil,
		CustomRootCommandHelpTemplate:   "",
		SliceFlagSeparator:              "",
		DisableSliceFlagSeparator:       false,
		MapFlagKeyValueSeparator:        "",
		UseShortOptionHandling:          false,
		Suggest:                         false,
		AllowExtFlags:                   false,
		SkipFlagParsing:                 false,
		CustomHelpTemplate:              "",
		PrefixMatchCommands:             false,
		SuggestCommandFunc:              nil,
		MutuallyExclusiveFlags:          []cli.MutuallyExclusiveFlags{},
		Arguments:                       []cli.Argument{},
		ReadArgsFromStdin:               false,
		StopOnNthArg:                    new(int),
	}
	if err := cmd.Run(context.TODO(), os.Args); err != nil {
		fmt.Println("program error:", err)
	}
}

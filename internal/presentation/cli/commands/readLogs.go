package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Galdoba/gslog"
	"github.com/Galdoba/gslog/internal/infrastructure/config"
	"github.com/Galdoba/gslog/internal/presentation/cli/flags"
	"github.com/urfave/cli/v3"
)

func ReadLogs(cfg config.Config) *cli.Command {
	cmd := &cli.Command{
		Name:                            "read",
		Aliases:                         []string{},
		Usage:                           "",
		UsageText:                       "",
		ArgsUsage:                       "",
		Version:                         "",
		Description:                     "",
		DefaultCommand:                  "",
		Category:                        "",
		Commands:                        []*cli.Command{},
		Flags:                           []cli.Flag{},
		HideHelp:                        false,
		HideHelpCommand:                 false,
		HideVersion:                     false,
		EnableShellCompletion:           false,
		ShellCompletionCommandName:      "",
		ShellComplete:                   nil,
		ConfigureShellCompletionCommand: nil,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			fmt.Println("me before!")
			return ctx, nil
		},
		After:                    nil,
		Action:                   readAction(cfg),
		CommandNotFound:          nil,
		OnUsageError:             nil,
		InvalidFlagAccessHandler: nil,
		Hidden:                   false,
		Authors:                  []any{},
		Copyright:                "",
		Reader:                   nil,
		Writer:                   nil,
		ErrWriter:                nil,
		ExitErrHandler:           nil,
		Metadata:                 map[string]interface{}{},
		ExtraInfo: func() map[string]string {
			panic("TODO")
		},
		CustomRootCommandHelpTemplate: "",
		SliceFlagSeparator:            "",
		DisableSliceFlagSeparator:     false,
		MapFlagKeyValueSeparator:      "",
		UseShortOptionHandling:        false,
		Suggest:                       false,
		AllowExtFlags:                 false,
		SkipFlagParsing:               false,
		CustomHelpTemplate:            "",
		PrefixMatchCommands:           false,
		SuggestCommandFunc:            nil,
		MutuallyExclusiveFlags:        []cli.MutuallyExclusiveFlags{},
		Arguments:                     []cli.Argument{},
		ReadArgsFromStdin:             false,
		StopOnNthArg:                  new(int),
	}
	return cmd
}

func readAction(cfg config.Config) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		fmt.Println("I view logs!")
		fmt.Println("args provided:")
		allPaths := []string{}
		switch len(c.Args().Slice()) {
		case 0:
			fmt.Println("nothing")
			fmt.Println("search working directory...")
			allPaths = append(allPaths, ".")
		default:
		}
		files, err := collectFiles(c)
		if err != nil {
			return fmt.Errorf("failed to collect files: %w", err)
		}
		entries := []string{}

		for _, file := range files {
			f, _ := os.Open(file)
			localEntries := gslog.ExtractStructures(f, nil)
			fmt.Println(len(localEntries))
			entries = append(entries, localEntries...)

		}

		fmt.Println("config:", cfg)
		for _, e := range entries {
			fmt.Println(e)
		}

		return nil
	}
}

func collectFiles(c *cli.Command) ([]string, error) {
	strictCheck := c.Bool(flags.STRICT)
	files := []string{}
	provided := c.Args().Slice()
	if len(provided) == 0 {
		provided = append(provided, ".")
	}
	for _, arg := range provided {
		f, err := os.Stat(arg)
		if err != nil {
			switch strictCheck {
			case true:
				gslog.Error("file descriptor retrival failed", "arg", arg, "error", err)
				return nil, fmt.Errorf("failed to handle argument %v: %w", arg, err)
			case false:
				gslog.Warn("file descriptor retrival failed", "arg", arg, "error", err)
				continue
			}
		}
		path, err := filepath.Abs(arg)
		if err != nil {
			switch strictCheck {
			case true:
				gslog.Error("absolute path retrival failed", "arg", arg, "error", err)
				return nil, fmt.Errorf("failed to handle argument %v: %w", arg, err)
			case false:
				gslog.Warn("absolute path retrival failed", "arg", arg, "error", err)
				continue
			}
		}
		switch f.IsDir() {
		case false:
			files = append(files, path)
		case true:
			fi, err := os.ReadDir(path)
			if err != nil {
				switch strictCheck {
				case true:
					gslog.Error("read directory", "arg", arg, "status", "failed")
					return nil, fmt.Errorf("failed to handle argument %v: %w", arg, err)
				case false:
					gslog.Warn("read directory", "arg", arg, "status", "failed")
					continue
				}
			}
			for _, f := range fi {
				if !f.IsDir() {
					files = append(files, filepath.Join(path, f.Name()))
				}
			}
		}
	}
	return files, nil
}

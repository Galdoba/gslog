package actions

import (
	"context"
	"fmt"

	"github.com/Galdoba/gslog/internal/infrastructure/config"
	"github.com/urfave/cli/v3"
)

func View(cfg config.Config) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		fmt.Println("I view logs!")
		fmt.Println(c.Root().Flags[0].Names())
		fmt.Println(c.Root().Flags[0].Get())
		fmt.Println(c.Root().String("d"))
		return nil
	}
}

package flags

import (
	"github.com/urfave/cli/v3"
)

const (
	DIR = "dir"
)

var DirFlag = &cli.StringFlag{
	Name:        DIR,
	HideDefault: true,
	Usage:       "set directory to serch logs in",
	Value:       ".",
	Aliases:     []string{"d"},
	Category:    "",
	DefaultText: "",
	Required:    true,
	Hidden:      false,
	Destination: new(string),
	Config:      cli.StringConfig{},
	OnlyOnce:    false,

	ValidateDefaults: false,
}

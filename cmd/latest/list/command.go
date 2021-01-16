package list

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/IPA-CyberLab/latest/pkg/fetch"
	"github.com/IPA-CyberLab/latest/pkg/parser"
)

var Command = &cli.Command{
	Name:  "list",
	Usage: "List all releases of the specified software",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "semver",
			Usage: "display parsed semver instead of the original version name.",
		},
	},
	Action: func(c *cli.Context) error {
		q, err := parser.Parse(c.Args().First())
		if err != nil {
			return err
		}

		fetcher := fetch.Direct{}
		rs, err := q.Execute(c.Context, fetcher)
		if err != nil {
			return err
		}

		if c.Bool("semver") {
			for _, r := range rs {
				fmt.Printf("%s\n", r.Version)
			}

			return nil
		}
		for _, r := range rs {
			fmt.Printf("%s\n", r.OriginalName)
		}

		return nil
	},
}

package list

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/urfave/cli/v2"

	"github.com/IPA-CyberLab/latest/pkg/fetch"
)

var Command = &cli.Command{
	Name:  "list",
	Usage: "List all releases of the specified software",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "range",
			Usage: "semver `RANGE` to query from. See https://godoc.org/github.com/blang/semver/v4#ParseRange for the accepted syntax.",
		},
		&cli.BoolFlag{
			Name:  "semver",
			Usage: "display parsed semver instead of the original version name.",
		},
	},
	Action: func(c *cli.Context) error {
		var vrange semver.Range
		vrangestr := c.String("range")
		if vrangestr != "" {
			var err error
			vrange, err = semver.ParseRange(vrangestr)
			if err != nil {
				return err
			}
		}

		softwareId := strings.TrimSpace(c.Args().First())
		rs, err := fetch.Direct{}.Fetch(c.Context, softwareId)
		if err != nil {
			return err
		}
		rs = rs.SelectAll(vrange)

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

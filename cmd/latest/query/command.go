package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/IPA-CyberLab/latest/pkg/fetch"
	"github.com/IPA-CyberLab/latest/pkg/parser"
	"github.com/IPA-CyberLab/latest/pkg/releases"
)

type AssetQueryType int

const (
	AssetQueryNone AssetQueryType = iota
	AssetQueryAll
	AssetQueryGuess
)

type OutputType int

const (
	OutputTypeLine OutputType = iota
	OutputTypeJson
)

var Command = &cli.Command{
	Name:  "query",
	Usage: "Query a release of specified software",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "asset",
			Usage: "Output asset URL. `QUERY` can be \"all\" to output all found assets, or specify \"guess\" to guess the most suitable one.",
			Value: "guess",
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "Output format. `FORMAT` can be \"line\" or \"json\".",
			Value: "line",
		},
	},
	Action: func(c *cli.Context) error {
		assetQ := AssetQueryNone
		if c.IsSet("asset") {
			s := c.String("asset")
			switch s {
			case "all":
				assetQ = AssetQueryAll
			case "guess":
				assetQ = AssetQueryGuess
			default:
				return fmt.Errorf("Unknown asset option: %q", s)
			}
		}

		outputType := OutputTypeLine
		if c.IsSet("output") {
			s := c.String("output")
			switch s {
			case "line":
				outputType = OutputTypeLine
			case "json":
				outputType = OutputTypeJson
			default:
				return fmt.Errorf("Unknown output type: %q", s)
			}
		}

		q, err := parser.Parse(c.Args().First())
		if err != nil {
			return err
		}

		fetcher := fetch.Direct{}
		rs, err := q.Execute(c.Context, fetcher)
		if err != nil {
			return err
		}
		if len(rs) == 0 {
			return errors.New("No release matched.")
		}
		r := rs[0]

		if assetQ == AssetQueryNone && outputType != OutputTypeLine {
			assetQ = AssetQueryGuess
		}
		switch assetQ {
		case AssetQueryNone:
			break
		case AssetQueryAll:
			if len(r.AssetURLs) == 0 {
				return releases.NoAssetFoundErr{OriginalName: r.OriginalName}
			}
		case AssetQueryGuess:
			pick, err := r.PickAsset()
			if err != nil {
				return err
			}
			r.AssetURLs = []string{pick}
		}

		switch outputType {
		case OutputTypeLine:
			if assetQ != AssetQueryNone {
				for _, u := range r.AssetURLs {
					fmt.Printf("%s\n", u)
				}
				return nil
			}
			fmt.Printf("%s\n", r.OriginalName)

		case OutputTypeJson:
			bs, err := json.MarshalIndent(r, "", "\t")
			if err != nil {
				return err
			}

			if _, err := os.Stdout.Write(bs); err != nil {
				return err
			}

		default:
			panic("not reached.")
		}
		return nil
	},
}

package app

import (
	"fmt"
	"io"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/IPA-CyberLab/latest/cmd/latest/list"
	"github.com/IPA-CyberLab/latest/cmd/latest/query"
	"github.com/IPA-CyberLab/latest/cmd/latest/serve"
	"github.com/IPA-CyberLab/latest/version"
)

func New() *cli.App {
	app := cli.NewApp()
	app.Name = "latest"
	app.Usage = "Query latest release versions"
	app.Authors = []*cli.Author{
		{Name: "yzp0n", Email: "yzp0n@coe.ad.jp"},
	}
	app.Version = fmt.Sprintf("%s.%s", version.Version, version.Commit)
	app.Commands = []*cli.Command{
		query.Command,
		list.Command,
		serve.Command,
	}
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "log-location",
			Usage: "Annotate logs with code location where the log was output",
		},
		&cli.BoolFlag{
			Name:  "log-json",
			Usage: "Format logs in json",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Enable verbose output",
		},
	}
	BeforeImpl := func(c *cli.Context) error {
		var logger *zap.Logger
		if loggeri, ok := app.Metadata["Logger"]; ok {
			logger = loggeri.(*zap.Logger)
		} else {
			cfg := zap.NewProductionConfig()
			cfg.DisableCaller = !c.Bool("log-location")
			if !c.Bool("log-json") {
				cfg.Encoding = "console"

				cfg.EncoderConfig.TimeKey = ""
				cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			}
			if c.Bool("verbose") {
				cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
			}

			var err error
			logger, err = cfg.Build(
				zap.AddStacktrace(zap.NewAtomicLevelAt(zap.DPanicLevel)))
			if err != nil {
				return err
			}
		}

		zap.ReplaceGlobals(logger)

		return nil
	}
	app.Before = func(c *cli.Context) error {
		if err := BeforeImpl(c); err != nil {
			// Print error message to stderr
			app.Writer = app.ErrWriter

			// Suppress help message on app.Before() failure.
			cli.HelpPrinter = func(_ io.Writer, _ string, _ interface{}) {}
			return err
		}

		return nil
	}
	app.After = func(c *cli.Context) error {
		_ = zap.L().Sync()
		return nil
	}

	return app
}

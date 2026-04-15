package main

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/iagafon/worker-service/cmd"
	"github.com/iagafon/worker-service/internal/pkg/constant"
)

func main() {
	const Usage = "Worker Service для обработки событий"

	const Description = `
Worker Service обрабатывает события из Kafka:
- ORDER_CREATED: расчёт стоимости доставки

Доступные команды:
  ./worker-service web-server              - HTTP сервер
  ./worker-service subscribe-order-created - Consumer для ORDER_CREATED
`

	app := cli.App{
		Name:    constant.AppName,
		Version: constant.GetFullVersion(),
		Usage:   Usage,
		Commands: []*cli.Command{
			cmd.WebServer(),
			cmd.SubscribeOrderCreated(),
		},
		Description: strings.TrimSpace(Description),
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "Человеко-читаемый формат для логов вместо JSON",
			},
		},
		EnableBashCompletion: true,
		CommandNotFound: func(cCtx *cli.Context, command string) {
			_ = cli.ShowAppHelp(cCtx)
		},
		HideHelpCommand: true,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}

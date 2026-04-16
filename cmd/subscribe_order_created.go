package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/iagafon/worker-service/internal/app/builder"
)

const (
	cmdSubscribeOrderCreatedUsage = "Подписка на топик событий создания заказов"

	cmdSubscribeOrderCreatedDescription = `
ВНИМАНИЕ!
Это команда подписки! Вы НЕ МОЖЕТЕ запустить эту команду в скрипт режиме.
Эта команда запускается только в режиме демона и работает до тех пор,
пока вы принудительно ее не остановите.

Команда подписывается на топик order.created для получения информации
о созданных заказах. Позволяет обрабатывать события и выполнять
дополнительные бизнес-операции (расчёт доставки, уведомления и т.д.).
`
)

func SubscribeOrderCreated() *cli.Command {
	return &cli.Command{
		Name:            "subscribe-order-created",
		Aliases:         []string{"suborder"},
		Usage:           cmdSubscribeOrderCreatedUsage,
		Description:     strings.TrimSpace(cmdSubscribeOrderCreatedDescription),
		Action:          cmdSubscribeOrderCreated,
		HideHelpCommand: true,
	}
}

func cmdSubscribeOrderCreated(cCtx *cli.Context) error {
	app := builder.NewBuilder(cCtx)
	app.BuildConfig()

	app.BuildRepoConnRedis()
	app.BuildBrokerKafka()

	app.BuildModuleClient()

	app.BuildRepoCurrencyRate()

	app.BuildModuleCurrency()
	app.BuildModuleDelivery()

	app.BuildHandlerEventOrder()

	app.BuildProcEventSubscribeOrderCreated()

	app.Run()
	return nil
}

package eprocessor

import (
	"context"
	"sync"

	"github.com/iagafon/pkg-broker"
	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/entity"
	ehandler "github.com/iagafon/worker-service/internal/app/handler/event"
	"github.com/iagafon/worker-service/internal/app/processor"
)

type orderCreatedProc struct {
	h   ehandler.Order
	bus broker.Bus[entity.EventOrderCreated]
}

func NewOrderCreatedEventsCatcher(
	h ehandler.Order,
	bus broker.Bus[entity.EventOrderCreated],
) processor.Processor {
	return &orderCreatedProc{h, bus}
}

func (p *orderCreatedProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	const S1 = "Не удалось подписаться на топик"
	const S2 = "Подписка на топик инициализирована"

	if err := p.bus.Subscribe(ctx, wg, p.h.CallbackOrderCreated); err != nil {
		panic(S1)
	}
	log.Debug().Msg(S2)
}

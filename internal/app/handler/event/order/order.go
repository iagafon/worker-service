package order

import (
	"context"
	"fmt"

	butil "github.com/iagafon/pkg-broker/util"
	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/entity"
	ehandler "github.com/iagafon/worker-service/internal/app/handler/event"
	"github.com/iagafon/worker-service/internal/pkg/http/binding"
)

type handler struct{}

func NewHandler() ehandler.Order {
	return &handler{}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context,
	event *entity.EventOrderCreated,
	headers map[string]string,
) error {
	log.Info().
		Ctx(ctx).
		Any("msg_body", event).
		Any("msg_headers", headers).
		Msg("Получено событие ORDER_CREATED")

	if err := binding.OnlyValidate(event); err != nil {
		return butil.NotCriticalError(fmt.Errorf("невалидные данные в EventOrderCreated: %w", err))
	}

	return nil
}

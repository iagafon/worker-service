package order

import (
	"context"
	"fmt"

	broker "github.com/iagafon/pkg-broker"
	butil "github.com/iagafon/pkg-broker/util"
	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/entity"
	ehandler "github.com/iagafon/worker-service/internal/app/handler/event"
	"github.com/iagafon/worker-service/internal/app/service"
	"github.com/iagafon/worker-service/internal/pkg/http/binding"
)

type handler struct {
	deliveryService            service.Delivery
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewHandler(
	deliveryService service.Delivery,
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated],
) ehandler.Order {
	return &handler{
		deliveryService:            deliveryService,
		busOrderDeliveryCalculated: busOrderDeliveryCalculated,
	}
}

func (h *handler) CallbackOrderCreated(
	ctx context.Context,
	event *entity.EventOrderCreated,
	headers map[string]string,
) error {
	log.Info().
		Ctx(ctx).
		Str("order_id", event.OrderID).
		Str("currency", event.Currency).
		Float64("total_amount", event.TotalAmount).
		Msg("Получено событие ORDER_CREATED")

	if err := binding.OnlyValidate(event); err != nil {
		return butil.NotCriticalError(fmt.Errorf("невалидные данные в EventOrderCreated: %w", err))
	}

	deliveryEvent, err := h.deliveryService.CalculateDeliveryPrice(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to calculate delivery price: %w", err)
	}

	// Отправка события ORDER_DELIVERY_CALCULATED
	if err = h.busOrderDeliveryCalculated.Send(ctx, deliveryEvent); err != nil {
		return fmt.Errorf("failed to send ORDER_DELIVERY_CALCULATED: %w", err)
	}

	log.Info().
		Ctx(ctx).
		Str("order_id", event.OrderID).
		Float64("delivery_price", deliveryEvent.DeliveryPrice).
		Str("currency", deliveryEvent.Currency).
		Msg("Событие ORDER_DELIVERY_CALCULATED отправлено")

	return nil
}

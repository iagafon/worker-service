package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	broker "github.com/iagafon/pkg-broker"
	"github.com/iagafon/pkg-broker/codec"
	putil "github.com/iagafon/pkg-broker/util"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/iagafon/worker-service/internal/app/client/fixer"
	"github.com/iagafon/worker-service/internal/app/config"
	"github.com/iagafon/worker-service/internal/app/config/section"
	"github.com/iagafon/worker-service/internal/app/entity"
	ehandler "github.com/iagafon/worker-service/internal/app/handler/event"
	"github.com/iagafon/worker-service/internal/app/handler/event/order"
	rhandler "github.com/iagafon/worker-service/internal/app/handler/http"
	"github.com/iagafon/worker-service/internal/app/handler/http/example"
	"github.com/iagafon/worker-service/internal/app/processor"
	eprocessor "github.com/iagafon/worker-service/internal/app/processor/event"
	rprocessor "github.com/iagafon/worker-service/internal/app/processor/http"
	pprocessor "github.com/iagafon/worker-service/internal/app/processor/other"
	"github.com/iagafon/worker-service/internal/app/repository"
	rcpostgres "github.com/iagafon/worker-service/internal/app/repository/conn/postgres"
	rcredis "github.com/iagafon/worker-service/internal/app/repository/conn/redis"
	rcurrency "github.com/iagafon/worker-service/internal/app/repository/currency"
	"github.com/iagafon/worker-service/internal/app/service"
	scurrency "github.com/iagafon/worker-service/internal/app/service/currency"
	sdelivery "github.com/iagafon/worker-service/internal/app/service/delivery"
	"github.com/iagafon/worker-service/internal/pkg/http/httph"
)

// Builder — структура для сборки зависимостей приложения.
// Использует паттерн Builder для последовательной инициализации компонентов.
type Builder struct {
	cCtx     *cli.Context
	ctx      context.Context
	wg       sync.WaitGroup
	err      error
	chErrors chan error

	cfg *config.Config

	// Подключения
	connPostgres *rcpostgres.Client
	connRedis    *rcredis.Client
	fixerClient  *fixer.Client

	brokerKafka                *broker.KafkaClient
	busOrderCreated            broker.Bus[entity.EventOrderCreated]
	busOrderDeliveryCalculated broker.Bus[entity.EventOrderDeliveryCalculated]

	// Процессоры
	processors []processor.Processor

	// HTTP middleware (OpenTelemetry, NewRelic, и др.)
	middlewares []httph.Middleware

	// Репозитории
	repoCurrencyRate repository.CurrencyRate

	// Handlers
	hExample rhandler.Example

	handlerEventOrder ehandler.Order

	// Модули
	moduleCurrency service.Currency
	moduleDelivery service.Delivery
}

// NewBuilder создаёт новый Builder и настраивает обработку сигналов OS.
// При получении SIGINT/SIGTERM контекст будет отменён.
func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{cCtx: cCtx, chErrors: make(chan error, 4096)} // <- добавить chErrors
	var cancelFunc func()
	b.ctx, cancelFunc = context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(sig, cancelFunc)
	go b.printErrors() // <- добавить

	return &b
}

////////////////////////////////////////////////////////////////////////////////
///// CONFIG ///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildConfig загружает конфигурацию из .env и переменных окружения.
// Можно передать injectors для модификации конфига после загрузки.
func (b *Builder) BuildConfig(injectors ...func(c *config.Config)) {
	b.buildConfig(config.LoadArgs{}, injectors)
}

// BuildConfigSimple загружает конфиг без файла .env (только injectors).
func (b *Builder) BuildConfigSimple(injectors ...func(c *config.Config)) {
	b.buildConfig(config.LoadArgs{SkipConfig: true}, injectors)
}

func (b *Builder) buildConfig(args config.LoadArgs, injectors []func(c *config.Config)) {
	if b.err != nil {
		return
	}

	// Определяем формат логов из CLI флага
	if b.cCtx != nil && b.cCtx.Bool("no-json") {
		args.EnableSimpleLog = true
	}
	args.Output = os.Stdout

	config.Load(args)

	// Применяем injectors
	for _, injector := range injectors {
		if injector != nil {
			injector(&config.Root)
		}
	}

	b.cfg = &config.Root
}

////////////////////////////////////////////////////////////////////////////////
///// BROKER ///////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildBrokerKafka() {
	b.exec(true, (*Builder).buildBrokerKafka)
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildRepoConnPostgres инициализирует подключение к PostgreSQL.
func (b *Builder) BuildRepoConnPostgres() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Repository.Postgres

		var err error
		b.connPostgres, err = rcpostgres.NewConn(b.ctx, cfg)
		if err != nil {
			b.err = fmt.Errorf("Repo.Conn.Postgres: %w", err)
			return
		}

		log.Debug().Msg("Unit Repo.Conn.Postgres has been initialized")
	})
}

// BuildRepoConnMigrator добавляет процессор миграций.
func (b *Builder) BuildRepoConnMigrator() {
	b.exec(b.connPostgres != nil, func(b *Builder) {
		proc := pprocessor.NewMigrator(b.connPostgres)
		b.processors = append(b.processors, proc)
	})
}

func (b *Builder) BuildRepoConnRedis() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Repository.Redis
		b.connRedis, b.err = rcredis.NewConn(b.ctx, cfg)
	})
}

// BuildRepoCurrencyRate initializes the currency rate repository.
func (b *Builder) BuildRepoCurrencyRate() {
	b.exec(true, (*Builder).buildRepoCurrencyRate, b.connRedis)
}

func (b *Builder) buildRepoCurrencyRate() {
	cfg := b.cfg.Client.Fixer
	b.repoCurrencyRate = rcurrency.NewRedisRepository(b.connRedis, cfg.CacheTTL)
	log.Info().Msg("Currency rate repository created")
}

////////////////////////////////////////////////////////////////////////////////
///// MODULES //////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildModuleClient() {
	b.exec(true, func(b *Builder) {
		b.fixerClient = fixer.NewClient(section.ClientFixer{
			ApiKey:  b.cfg.Client.Fixer.ApiKey,
			BaseURL: b.cfg.Client.Fixer.BaseURL,
		})
	})
}

// BuildModuleCurrency initializes currency service.
func (b *Builder) BuildModuleCurrency() {
	b.exec(true, (*Builder).buildModuleCurrency, b.fixerClient, b.repoCurrencyRate)
}

func (b *Builder) buildModuleCurrency() {
	b.moduleCurrency = scurrency.NewService(b.fixerClient, b.repoCurrencyRate)
	log.Info().Msg("Currency service created")
}

func (b *Builder) BuildModuleDelivery() {
	b.exec(true, (*Builder).buildModuleDelivery, b.moduleCurrency)
}

func (b *Builder) buildModuleDelivery() {
	b.moduleDelivery = sdelivery.NewService(b.moduleCurrency)
	log.Info().Msg("Delivery service created")
}

////////////////////////////////////////////////////////////////////////////////
///// HANDLERS /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildHandlerExample создаёт example handler.
func (b *Builder) BuildHandlerExample() {
	b.exec(true, func(b *Builder) {
		b.hExample = example.NewHandler()
		log.Debug().Msg("Unit Handler.Example has been initialized")
	})
}

func (b *Builder) BuildHandlerEventOrder() {
	b.exec(true, func(b *Builder) {
		b.handlerEventOrder = order.NewHandler(b.moduleDelivery, b.busOrderDeliveryCalculated)
	}, b.moduleDelivery, b.busOrderDeliveryCalculated)
}

// TODO: Добавить методы для других handlers:
// func (b *Builder) BuildHandlerUser() { ... }

////////////////////////////////////////////////////////////////////////////////
///// PROCESSORS ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// BuildProcHttp создаёт и добавляет HTTP процессор.
func (b *Builder) BuildProcHttp() {
	b.exec(true, func(b *Builder) {
		cfg := config.Root.Processor.WebServer
		proc := rprocessor.NewHTTP(b.hExample, b.middlewares, cfg)
		b.processors = append(b.processors, proc)
	})
}

func (b *Builder) BuildProcEventSubscribeOrderCreated() {
	b.exec(true, (*Builder).buildProcEventSubscribeOrderCreated, b.handlerEventOrder, b.busOrderCreated)
}

////////////////////////////////////////////////////////////////////////////////
///// RUN //////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// Run запускает все подготовленные процессоры и ожидает их завершения.
func (b *Builder) Run() {
	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Ошибка при инициализации приложения")
	}

	log.Info().Msg("Приложение инициализировано")
	defer log.Info().Msg("Приложение завершено, до свидания!")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// INTERNAL /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) buildProcEventSubscribeOrderCreated() {
	proc := eprocessor.NewOrderCreatedEventsCatcher(b.handlerEventOrder, b.busOrderCreated)
	b.processors = append(b.processors, proc)
	log.Info().Msg("Processor ORDER_CREATED registered")
}

func (b *Builder) buildBrokerKafka() {
	kafkaConfig := broker.KafkaConfig{
		Addresses:     b.cfg.Broker.Kafka.Addresses,
		ConsumerGroup: b.cfg.Broker.Kafka.ConsumerGroup,
		ClientID:      b.cfg.Broker.Kafka.ClientID,
	}
	log.Debug().
		Any("addresses", kafkaConfig.Addresses).
		Str("group", b.cfg.Broker.Kafka.ConsumerGroup).
		Msg("kafka config")

	var err error

	if b.brokerKafka, err = broker.NewKafkaClient(kafkaConfig); err != nil {
		b.err = fmt.Errorf("failed to create kafka client: %w", err)
		return
	}

	type T1 = entity.EventOrderCreated
	type T2 = entity.EventOrderDeliveryCalculated

	t1codec := codec.NewCodecJson[T1]()
	t2codec := codec.NewCodecJson[T2]()

	b.busOrderCreated = broker.MustKafkaBus(
		b.brokerKafka,
		t1codec,
		b.cfg.Broker.Kafka.ModelOrder.Created.Topic,
		putil.Coalesce(b.cfg.Broker.Kafka.ConsumerGroup, b.cfg.Broker.Kafka.ModelOrder.Created.ConsumerGroup),
	)

	b.busOrderDeliveryCalculated = broker.MustKafkaBus(
		b.brokerKafka,
		t2codec,
		b.cfg.Broker.Kafka.ModelOrder.DeliveryCalculated.Topic,
		putil.Coalesce(b.cfg.Broker.Kafka.ConsumerGroup, b.cfg.Broker.Kafka.ModelOrder.Created.ConsumerGroup),
	)
	log.Info().Msg("Kafka buses created")
}

// waitForSignal ожидает сигнал и вызывает cancelFunc.
func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	gotSig := <-sig
	log.Info().Str("sig", gotSig.String()).Msg("Запрошено завершение")
	cancelFunc()
}

func (b *Builder) exec(preCond bool, cb func(b *Builder), requiredArgs ...any) {
	if !preCond || b.err != nil || b.ctx.Err() != nil {
		return
	}

	for _, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}

		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}

func (b *Builder) printErrors() {
	for err := range b.chErrors {
		log.Error().Err(err).Msg("Got new error")
	}
}

# pkg-broker

Go библиотека для работы с брокером сообщений (Kafka) в микросервисной архитектуре.

---

## 📖 Описание

Библиотека предоставляет удобную абстракцию `Bus`, основанную на [Sarama](https://github.com/IBM/sarama), для отправки и получения сообщений через Kafka. 
Используется в качестве сторонней зависимости для микросервисов `order-service` и `worker-service`.

- Версия Go 1.25
- Kafka 3.7.0

#### Releases:
- [1.0.0](https://github.com/iagafon/pkg-broker)

---

## ⚙️ Основные компоненты

- **Bus** - интерфейс для работы с топиками Kafka
- **KafkaClient** - клиент для подключения к Kafka
- **Codec** - система кодирования/декодирования сообщений
- **MessageHandler** - обработчики входящих сообщений

---

## 🚀 Быстрый старт

Перед установкой следует выполнить следующие шаги

1) Сообщим Go, что это приватный модуль

    ```bash
    go env -w GOPRIVATE=github.com/iagafon/*
    go env -w GONOSUMDB=github.com/iagafon/*
    go env -w GONOPROXY=github.com/iagafon/*
    ```

2) Настроить аутентификацию Git

- Проверить аутентификацию Git:

    ```bash
    ssh -T git@github.com
    ```

- Установить авторизацию по ssh 

    ```bash
    git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
    ```
- Скачать пакет

    ```bash
    go get github.com/iagafon/pkg-broker
    ```

- Examples

    ```bash
    go test -v -run ExampleBusMock
    ```

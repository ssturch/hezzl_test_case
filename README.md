# hezzl_test_case
## Задача:
 - Развернуть сервис на Golang, Postgres, Clickhouse, Nats (альтернатива kafka), Redis
## Условие: 
 - согласно описанию в файле ***HEZZL TASK.pdf***
## Результат:
### Создано 2 сервиса:
 - ***cmd/hezzlapi/main.go*** - сервис, являющийся REST-API сервером, работающий c PostgreSQL, Redis, Nats.
 - ***cmd/ntstoclcks/main.go*** - сервис, получающий от REST-API сервера логи с помощью Nats и отправляющий их пачкой в ClickHouse.
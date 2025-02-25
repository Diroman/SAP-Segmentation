# SAP Segmentation Importer

## Описание

`SAP Segmentation Importer` — это Go-приложение для импорта данных сегментации из сторонней ERP-системы в базу данных
PostgreSQL. Проект включает:

- Импорт данных через HTTP API с пагинацией.
- Сохранение данных в таблицу `segmentation` с обработкой конфликтов (upsert).
- Логирование в консоль и файл.
- Docker Compose для поднятия приложения, mock ERP-сервера и PostgreSQL.

## Требования

- Go 1.21+
- Docker и Docker Compose
- Make (для сборки через Makefile)

## Запуск

### Локально

1. Соберите приложение:
    ```bash
    make build
    ```

2. Запустите:
    ``` bash
    make run
    ```

Убедитесь, что PostgreSQL и ERP-сервер доступны по настройкам в переменных окружения.

### Через Docker Compose

1. Запустите все сервисы:
   ```bash
   docker-compose up --build
   ```
   Это поднимет:
   app (sap_segmentation)
   erp (mock ERP-сервер на :8080)
   db (PostgreSQL на :5432)
2. Остановите:

   ```bash
   docker-compose down
   ```

   Для удаления данных базы:
   ```bash
   docker-compose down --volumes
   ```

## Конфигурация

Переменные окружения задаются в `docker-compose.yml`. Основные параметры:

`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` — подключение к PostgreSQL.  
`CONN_URI` — URL внешнего API (по умолчанию `http://erp:8080/ords/bsm/segmentation/get_segmentation`).   
`IMPORT_BATCH_SIZE` — размер пачки данных (по умолчанию `50`).  
`LOG_LEVEL` — уровень логирования (например, `debug`, `info`).

## Использование

После запуска приложение автоматически:

1. Подключается к PostgreSQL (db).
2. Запрашивает данные из ERP (erp) с пагинацией.
3. Сохраняет данные в таблицу segmentation.

Логи записываются в /logs/segmentation_import.log.

## Тестирование

### Проверка ERP:

```bash
curl -H "Authorization: Basic 4Dfddf5:jKlljHGH" "http://localhost:8080/ords/bsm/segmentation/get_segmentation?p_limit=10&p_offset=0"
```

### Проверка БД:

```bash
docker-compose exec db psql -U postgres -d mesh_group -c "SELECT * FROM segmentation LIMIT 5;"
```
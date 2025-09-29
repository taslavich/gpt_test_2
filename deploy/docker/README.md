# Образы микросервисов

В каталоге `deploy/docker` находятся Dockerfile для сборки всех основных сервисов:

- `bid-engine`
- `orchestrator`
- `router`
- `spp-adapter`
- `kafka-loader`
- `clickhouse-loader`

Скрипт `build.sh` использует эти файлы по умолчанию. Примеры:

```bash
# Собрать все образы и поместить их в локальный registry
./build.sh push-local

# Собрать конкретный сервис
./build.sh bid-engine
```

## GeoIP база для SPP Adapter

Для работы `spp-adapter` требуется база GeoIP. Образ создаёт каталог `/data` и ожидает, что файл будет доступен по пути
`/data/GeoIP2_City.mmdb`. Его можно:

1. Подмонтировать при запуске контейнера:
   ```bash
   docker run -v /local/path/GeoIP2_City.mmdb:/data/GeoIP2_City.mmdb:ro \
     -e GEO_IP_DB_PATH=/data/GeoIP2_City.mmdb \
     localhost:5000/exchange/spp-adapter:latest
   ```
2. Загрузить в кластер Kubernetes как `Secret` с именем `geoip-db` (см. деплоймент `spp-adapter`).

> ℹ️ Файл MaxMind можно хранить прямо в корне репозитория (например, `GeoIP2_City.mmdb`), чтобы сборка и деплой работали офлайн.

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

Для работы `spp-adapter` требуется база GeoIP. Во время сборки в образ копируется файл `GeoIP2_City.mmdb`
из корня репозитория (или другой путь, переданный через build-аргумент `GEOIP_DB_FILE`). Внутри контейнера
он располагается по адресу `/var/lib/geoip/GeoIP2_City.mmdb`, а переменная окружения `GEO_IP_DB_PATH`
прописана в ConfigMap Kubernetes и по умолчанию указывает на этот путь.

> ℹ️ Убедитесь, что перед сборкой в корне проекта лежит актуальная база `GeoIP2_City.mmdb` (её можно хранить в git).
> При необходимости можно указать другой файл: `docker build --build-arg GEOIP_DB_FILE=GeoLite2-City.mmdb ...`.

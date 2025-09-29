# Kubernetes деплой RTB Exchange

Документ описывает порядок развертывания проекта в Kubernetes, настройку внешнего шлюза и подготовку DNS/доменов.

## Компоненты

- **Redis** – деплоймент + service (`redis-service`).
- **Kafka (KRaft)** – statefulset + headless service (`kafka-headless`) и клиентский service (`kafka-service`).
- **ClickHouse/Kafka loaders** – отдельные деплойменты с ClusterIP сервисами.
- **Микросервисы** – `bid-engine`, `orchestrator`, `router`, `spp-adapter`.
- **Gateway** – NGINX-балансировщик, который принимает внешние HTTP(S) вызовы и проксирует их в сервисы по портам/путям.
- **Ingress** – опциональный слой, если в кластере есть установленный Ingress Controller.

## Gateway (балансировщик)

Файлы:

- `configs/gateway-config.yaml` – конфигурация NGINX.
- `deployments/gateway-deployment.yaml` – деплоймент с 2 репликами и health-чеками.
- `services/gateway-service.yaml` – сервис типа `LoadBalancer` с набором портов: `80/443` (HTTP(S) роутинг), `8080`–`8085` (прямой доступ к сервисам).

### HTTP/HTTPS-маршрутизация

На портах `80` (HTTP) и `443` (HTTPS) доступны следующие префиксы:

| Префикс                | Целевой сервис |
|------------------------|----------------|
| `/bid-engine/`         | bid-engine     |
| `/orchestrator/`       | orchestrator   |
| `/router/`             | router         |
| `/spp-adapter/`        | spp-adapter    |
| `/kafka-loader/`       | kafka-loader   |
| `/clickhouse-loader/`  | clickhouse-loader |

Проверка работоспособности – `curl http://<domain>/healthz` или `curl https://<domain>/healthz -k` (если используется self-signed сертификат).

### Прямой доступ по портам

| Внешний порт | Назначение |
|--------------|-----------|
| `443`        | HTTPS для всех маршрутов |
| `8080`       | bid-engine |
| `8081`       | orchestrator |
| `8082`       | router |
| `8083`       | spp-adapter |
| `8084`       | clickhouse-loader |
| `8085`       | kafka-loader |

## Настройка домена

1. Выполните `./deploy/setup-domain.sh <domain>` – скрипт выведет IP или hostname балансировщика и подсказки.
2. Чтобы автоматически дописать `/etc/hosts`, используйте `./deploy/setup-domain.sh <domain> --apply` (потребуется `sudo`).
3. Для боевого DNS добавьте A/AAAA-запись у провайдера, указывая на полученный IP.

После обновления DNS проверьте доступность:

```bash
curl http://<domain>/healthz
curl http://<domain>:8083/health   # прямой доступ к SPP adapter
```

## Сценарии деплоя

Основной скрипт – `deploy.sh`. Ключевые команды:

```bash
./deploy.sh all        # Полный деплой
./deploy.sh services   # Только микросервисы
./deploy.sh gateway    # Только внешний шлюз
./deploy.sh test       # Проверка доступности через балансировщик
```

Скрипт автоматически применяет ConfigMap/Secret, ожидает readiness и выводит статус.

### GeoIP база для SPP Adapter

`spp-adapter` требует файл `GeoIP2_City.mmdb`. Чтобы деплой не падал на этапе rollout:

1. Скачайте базу MaxMind (GeoIP2 или GeoLite2) и положите её в корень репозитория рядом с `deploy.sh`. Файл можно хранить под версией git.
2. При необходимости можно использовать каталог `deploy/assets`, но поиск в первую очередь идёт в корне.
3. Если секрет `geoip-db` уже создан в кластере, дополнительных действий не требуется.

При отсутствии файла скрипт завершится с подсказкой и не станет запускать `spp-adapter`, чтобы избежать бесконечного ожидания готовности.

## Egress для Router

Файл `configs/router-egress-policy.yaml` задаёт `NetworkPolicy`, разрешающую `router` обращаться к внешним HTTP/HTTPS ресурсам (порт 80/443) и к DNS (порт 53). Если в кластере не используется контроллер сетевых политик, манифест не оказывает влияния, но обеспечивает совместимость с кластерами, где политики включены.

## HTTPS

TLS уже включён по умолчанию. В репозитории добавлен dev/self-signed сертификат (`secrets/gateway-tls-secret.yaml`), который применяет `deploy.sh`. Для production замените значения на боевой сертификат:

1. Подготовьте свои `tls.crt` и `tls.key`.
2. Создайте secret: `kubectl create secret tls gateway-tls --key tls.key --cert tls.crt -n exchange --dry-run=client -o yaml > deploy/k8s/secrets/gateway-tls-secret.yaml`.
3. Примените `./deploy.sh gateway` или `kubectl apply -f deploy/k8s/secrets/gateway-tls-secret.yaml`.
4. Перезапустите шлюз: `kubectl rollout restart deployment/gateway-deployment -n exchange`.

## Примечания

- Если в окружении уже есть Ingress Controller, `deploy.sh ingress` применит `ingress/ingress.yaml` для маршрутизации по домену `rtb.local`.
- Для локальных кластеров (k3s/Minikube) сервис типа `LoadBalancer` автоматически создаёт NodePort'ы (30080, 31080-31085), что позволяет тестировать балансировщик по IP узла.
- Скрипт `deploy.sh test` использует балансировщик и проверяет `/health` основных сервисов.

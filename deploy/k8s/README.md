# Kubernetes деплой RTB Exchange

Документ описывает порядок развертывания проекта в Kubernetes, настройку внешнего шлюза и подготовку DNS/доменов.

## Компоненты

- **Redis** – деплоймент + service (`redis-service`).
- **Kafka (KRaft)** – statefulset + headless service (`kafka-headless`) и клиентский service (`kafka-service`).
- **ClickHouse/Kafka loaders** – отдельные деплойменты с ClusterIP сервисами.
- **Микросервисы** – `bid-engine`, `orchestrator`, `router`, `spp-adapter`.
- **Gateway** – NGINX-балансировщик, который принимает внешние HTTP вызовы и проксирует их в сервисы по портам/путям.
- **Ingress** – опциональный слой, если в кластере есть установленный Ingress Controller.

## Gateway (балансировщик)

Файлы:

- `configs/gateway-config.yaml` – конфигурация NGINX.
- `deployments/gateway-deployment.yaml` – деплоймент с 2 репликами и health-чеками.
- `services/gateway-service.yaml` – сервис типа `LoadBalancer` с набором портов: `80` (HTTP роутинг), `8080`–`8085` (прямой доступ к сервисам).

### HTTP-маршрутизация

На порту `80` доступны следующие префиксы:

| Префикс                | Целевой сервис |
|------------------------|----------------|
| `/bid-engine/`         | bid-engine     |
| `/orchestrator/`       | orchestrator   |
| `/router/`             | router         |
| `/spp-adapter/`        | spp-adapter    |
| `/kafka-loader/`       | kafka-loader   |
| `/clickhouse-loader/`  | clickhouse-loader |

Проверка работоспособности – `curl http://<domain>/healthz`.

### Прямой доступ по портам

| Внешний порт | Назначение |
|--------------|-----------|
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

## Egress для Router

Файл `configs/router-egress-policy.yaml` задаёт `NetworkPolicy`, разрешающую `router` обращаться к внешним HTTP/HTTPS ресурсам (порт 80/443) и к DNS (порт 53). Если в кластере не используется контроллер сетевых политик, манифест не оказывает влияния, но обеспечивает совместимость с кластерами, где политики включены.

## HTTPS

Для включения TLS:

1. Создайте секрет с сертификатом: `kubectl create secret tls gateway-tls --key tls.key --cert tls.crt -n exchange`.
2. Добавьте блок `server { listen 443 ssl; ... }` в `gateway-config.yaml` и смонтируйте секрет в `gateway-deployment.yaml` (volume + volumeMount).
3. Перезапустите шлюз: `kubectl rollout restart deployment/gateway-deployment -n exchange`.

## Примечания

- Если в окружении уже есть Ingress Controller, `deploy.sh ingress` применит `ingress/ingress.yaml` для маршрутизации по домену `rtb.local`.
- Для локальных кластеров (k3s/Minikube) сервис типа `LoadBalancer` автоматически создаёт NodePort'ы (30080, 31080-31085), что позволяет тестировать балансировщик по IP узла.
- Скрипт `deploy.sh test` использует балансировщик и проверяет `/health` основных сервисов.

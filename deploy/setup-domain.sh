#!/bin/bash
set -euo pipefail

K8S_NAMESPACE="${K8S_NAMESPACE:-exchange}"
SERVICE_NAME="${SERVICE_NAME:-gateway-service}"

usage() {
  cat <<USAGE
Usage: $0 <domain> [ip] [--apply]

<domain>  - обязательное DNS-имя (например, rtb.example.com)
[ip]      - необязательный IP. Если не указан, будет извлечен из сервиса Kubernetes ${SERVICE_NAME}
--apply   - при указании строка с доменом автоматически добавляется в /etc/hosts (потребуется sudo)

Скрипт всегда выводит инструкцию для ручного добавления записи.
USAGE
}

if [ $# -lt 1 ]; then
  usage
  exit 1
fi

DOMAIN=$1
shift

IP=""
APPLY=false

for arg in "$@"; do
  case "$arg" in
    --apply)
      APPLY=true
      ;;
    *)
      if [ -z "$IP" ]; then
        IP="$arg"
      else
        echo "Неизвестный параметр: $arg" >&2
        usage
        exit 1
      fi
      ;;
  esac
  shift || true
  set -- "$@"
done

if [ -z "$IP" ]; then
  echo "🔍 Получаем адрес балансировщика из Kubernetes..." >&2
  if command -v kubectl >/dev/null 2>&1; then
    LB_IP=$(kubectl get svc "$SERVICE_NAME" -n "$K8S_NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
    if [ -z "$LB_IP" ] || [ "$LB_IP" = "<no value>" ]; then
      LB_HOSTNAME=$(kubectl get svc "$SERVICE_NAME" -n "$K8S_NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)
      if [ -n "$LB_HOSTNAME" ] && [ "$LB_HOSTNAME" != "<no value>" ]; then
        IP="$LB_HOSTNAME"
      fi
    else
      IP="$LB_IP"
    fi

    if [ -z "$IP" ]; then
      NODE_IP=$(kubectl get nodes -o wide | awk '/Ready/ {print $6; exit}')
      if [ -n "$NODE_IP" ]; then
        IP="$NODE_IP"
      fi
    fi
  fi
fi

if [ -z "$IP" ]; then
  echo "⚠️  Не удалось автоматически определить IP. Укажите его вручную вторым аргументом." >&2
  usage
  exit 1
fi

if [ "$APPLY" = true ]; then
  if ! command -v sudo >/dev/null 2>&1; then
    echo "❌ Для автоматического добавления записи требуется sudo" >&2
    exit 1
  fi

  echo "🛠️  Добавляем запись $IP $DOMAIN в /etc/hosts..." >&2
  TMP_FILE=$(mktemp)
  if [ -f /etc/hosts ]; then
    grep -v "[[:space:]]$DOMAIN$" /etc/hosts > "$TMP_FILE"
  fi
  echo "$IP $DOMAIN" >> "$TMP_FILE"
  sudo cp "$TMP_FILE" /etc/hosts
  rm -f "$TMP_FILE"
  echo "✅ /etc/hosts обновлен" >&2
fi

echo "=============================="
echo "Добавьте запись DNS или hosts:" >&2
echo "  $IP $DOMAIN"
echo
cat <<'HINT'
Советы:
  • Создайте A/AAAA запись у своего DNS-провайдера, указывающую на IP сервиса.
  • Для локального тестирования можно воспользоваться /etc/hosts (см. ключ --apply).
  • После обновления DNS проверьте доступность: curl http://DOMAIN/healthz
HINT

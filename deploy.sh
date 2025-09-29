#!/bin/bash

set -e  # Останавливаться при ошибках

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$SCRIPT_DIR/deploy/k8s"
ASSETS_DIR="$SCRIPT_DIR/deploy/assets"
NAMESPACE="exchange"

echo "=== RTB Exchange Deployment ==="

usage() {
    echo "Usage: $0 [all|configs|redis|kafka|clickhouse|loaders|services|gateway|ingress|status|logs|test|clean|destroy]"
    echo "  all         - Full deployment (default)"
    echo "  configs     - Apply only configs"
    echo "  redis       - Deploy only Redis"
    echo "  kafka       - Deploy only Kafka cluster"
    echo "  clickhouse  - Configure ClickHouse Cloud connection"
    echo "  loaders     - Deploy only Kafka and ClickHouse loaders"
    echo "  services    - Deploy only microservices"
    echo "  gateway     - Deploy only external gateway"
    echo "  ingress     - Deploy only ingress"
    echo "  status      - Check deployment status"
    echo "  logs        - Show logs"
    echo "  test        - Test endpoints"
    echo "  clean       - Remove all resources but keep namespace"
    echo "  destroy     - COMPLETELY remove everything including namespace"
}

# Функция для настройки k3s
setup_k3s_registry() {
    echo "🔧 Configuring k3s for local registry..."
    
    # Проверяем, работает ли k3s
    if ! systemctl is-active k3s >/dev/null 2>&1; then
        echo "ℹ️ k3s is not running, skipping k3s configuration"
        return 0
    fi
    
    # Проверяем, не настроен ли уже k3s
    if [ -f "/etc/rancher/k3s/registries.yaml" ]; then
        if grep -q "localhost:5000" /etc/rancher/k3s/registries.yaml; then
            echo "✅ k3s registry already configured"
            return 0
        fi
    fi
    
    # Создаем backup текущих конфигов
    local backup_dir="/tmp/rtb-k3s-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"
    
    echo "📦 Creating backup in $backup_dir..."
    
    # Backup существующих конфигов
    if [ -f "/etc/docker/daemon.json" ]; then
        cp /etc/docker/daemon.json "$backup_dir/docker-daemon-backup.json"
        echo "✅ Docker daemon config backed up"
    fi
    
    if [ -f "/etc/rancher/k3s/registries.yaml" ]; then
        cp /etc/rancher/k3s/registries.yaml "$backup_dir/k3s-registries-backup.yaml"
        echo "✅ k3s registries config backed up"
    fi
    
    # 1. Настраиваем Docker daemon
    echo "📝 Configuring Docker daemon..."
    sudo mkdir -p /etc/docker
    sudo tee /etc/docker/daemon.json <<EOF
{
  "insecure-registries": ["localhost:5000", "127.0.0.1:5000", "registry.local:5000"]
}
EOF
    
    # Перезапускаем Docker
    echo "🔄 Restarting Docker..."
    sudo systemctl restart docker
    sleep 3
    
    # 2. Настраиваем k3s
    echo "📝 Configuring k3s registry settings..."
    sudo mkdir -p /etc/rancher/k3s
    sudo tee /etc/rancher/k3s/registries.yaml <<EOF
mirrors:
  "localhost:5000":
    endpoint:
      - "http://localhost:5000"
  "127.0.0.1:5000":
    endpoint:
      - "http://127.0.0.1:5000"
EOF
    
    # Перезапускаем k3s
    echo "🔄 Restarting k3s..."
    sudo systemctl stop k3s
    sleep 5
    sudo systemctl start k3s
    
    # Ждем пока k3s запустится
    echo "⏳ Waiting for k3s to start..."
    for i in {1..30}; do
        if kubectl get nodes >/dev/null 2>&1; then
            echo "✅ k3s is running"
            break
        fi
        sleep 2
        if [ $i -eq 30 ]; then
            echo "❌ k3s failed to start within 60 seconds"
            return 1
        fi
    done
    
    echo "✅ k3s registry configuration completed successfully!"
    echo "📋 Backup files saved in: $backup_dir"
}

# Автоматическая настройка перед деплоем
auto_setup_before_deploy() {
    echo "🔧 Auto-setting up environment for deployment..."

    # Настраиваем k3s если он установлен
    setup_k3s_registry

    # Проверяем и запускаем registry
    if ! curl -s http://localhost:5000/v2/_catalog >/dev/null; then
        echo "🚀 Starting local registry..."
        local build_script="$SCRIPT_DIR/build.sh"
        if [ -f "$build_script" ]; then
            "$build_script" registry-start
        else
            echo "❌ build.sh not found. Please start registry manually."
            return 1
        fi
    else
        echo "✅ Local registry is running"
    fi
    
    # Проверяем, что образы существуют в registry
    echo "🔍 Checking if images are available in registry..."
    local images_missing=0
    local services=("bid-engine" "orchestrator" "router" "spp-adapter" "kafka-loader" "clickhouse-loader")
    
    for service in "${services[@]}"; do
        if ! curl -s http://localhost:5000/v2/exchange/$service/tags/list | grep -q "latest"; then
            echo "❌ Image for $service not found in registry"
            images_missing=1
        fi
    done
    
    if [ $images_missing -eq 1 ]; then
        echo "⚠️ Some images missing in registry. Please run './build.sh push-local' first."
        return 1
    fi
    
    echo "✅ All images available in registry"
}

# Функция очистки ресурсов
clean_resources() {
    echo "🧹 Cleaning all resources in namespace $NAMESPACE..."
    
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        kubectl delete all --all -n "$NAMESPACE"
        kubectl delete configmap,secret,ingress --all -n "$NAMESPACE"
        echo "✅ All resources cleaned in namespace $NAMESPACE"
    else
        echo "ℹ️ Namespace $NAMESPACE does not exist"
    fi
}

# Функция полного удаления
destroy_namespace() {
    echo "💥 COMPLETELY destroying namespace $NAMESPACE..."
    
    if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        kubectl delete namespace "$NAMESPACE"
        echo "✅ Namespace $NAMESPACE destroyed"
        
        # Также останавливаем registry
        local build_script="$SCRIPT_DIR/build.sh"
        if [ -f "$build_script" ]; then
            echo "🛑 Stopping registry..."
            "$build_script" registry-stop
        fi
    else
        echo "ℹ️ Namespace $NAMESPACE does not exist"
    fi
}

# Функция деплоя конфигов
deploy_configs() {
    echo "📄 Deploying configs..."

    if [ -f "$K8S_DIR/namespace.yaml" ]; then
        kubectl apply -f "$K8S_DIR/namespace.yaml"
    fi

    if [ -d "$K8S_DIR/configs" ]; then
        kubectl apply -f "$K8S_DIR/configs/"
        echo "✅ Configs deployed"
    else
        echo "❌ Configs directory not found: $K8S_DIR/configs/"
        return 1
    fi

    if [ -d "$K8S_DIR/secrets" ]; then
        kubectl apply -f "$K8S_DIR/secrets/"
        echo "✅ Secrets deployed"
    fi
}

# Функция деплоя Redis
deploy_redis() {
    echo "🔴 Deploying Redis..."

    local redis_files=(
        "$K8S_DIR/deployments/redis-deployment.yaml"
        "$K8S_DIR/services/redis-service.yaml"
    )

    for file in "${redis_files[@]}"; do
        if [ -f "$file" ]; then
            kubectl apply -f "$file"
        else
            echo "❌ Redis file not found: $file"
            return 1
        fi
    done

    echo "⏳ Waiting for Redis to be ready..."
    kubectl rollout status deployment/redis-deployment -n "$NAMESPACE" --timeout=120s
    echo "✅ Redis deployed and ready"
}

# Функция деплоя Kafka
deploy_kafka() {
    echo "📊 Deploying Kafka cluster..."

    local kafka_files=(
        "$K8S_DIR/services/kafka-service.yaml"
        "$K8S_DIR/deployments/kafka-deployment.yaml"
    )

    for file in "${kafka_files[@]}"; do
        if [ -f "$file" ]; then
            kubectl apply -f "$file"
            echo "✅ Applied: $(basename "$file")"
        else
            echo "❌ Kafka file not found: $file"
            return 1
        fi
    done

    echo "⏳ Waiting for Kafka to be ready..."
    kubectl rollout status statefulset/kafka -n "$NAMESPACE" --timeout=300s
    kubectl wait --for=condition=ready pod/kafka-0 -n "$NAMESPACE" --timeout=120s

    echo "✅ Kafka deployed"
}

# Функция настройки ClickHouse Cloud
setup_clickhouse_cloud() {
    echo "☁️ Configuring ClickHouse Cloud connection..."

    # Проверяем наличие секрета с данными ClickHouse Cloud
    if kubectl get secret clickhouse-cloud-secret -n $NAMESPACE >/dev/null 2>&1; then
        echo "✅ ClickHouse Cloud secret already exists"
        return 0
    fi

    echo "📝 Please provide ClickHouse Cloud connection details:"
    read -p "ClickHouse Cloud Host: " ch_host
    read -p "ClickHouse Cloud Port (default: 9440): " ch_port
    read -p "ClickHouse Cloud Username: " ch_username
    read -s -p "ClickHouse Cloud Password: " ch_password
    echo
    read -p "ClickHouse Cloud Database (default: default): " ch_database

    ch_port=${ch_port:-9440}
    ch_database=${ch_database:-default}

    # Создаем секрет с данными подключения
    kubectl create secret generic clickhouse-cloud-secret \
        --namespace "$NAMESPACE" \
        --from-literal=host="$ch_host" \
        --from-literal=port="$ch_port" \
        --from-literal=username="$ch_username" \
        --from-literal=password="$ch_password" \
        --from-literal=database="$ch_database" \
        --dry-run=client -o yaml | kubectl apply -f -

    echo "✅ ClickHouse Cloud configuration saved as secret"

    # Обновляем конфиг clickhouse-loader для использования Cloud
    if [ -f "$K8S_DIR/configs/clickhouse-loader-config.yaml" ]; then
        echo "🔧 Updating clickhouse-loader config for Cloud..."
        # Создаем патч для использования Cloud соединения
        kubectl patch configmap clickhouse-loader-config -n "$NAMESPACE" --type merge \
            -p "{\"data\":{\"CLICK_HOUSE_DSN\":\"https://${ch_host}:${ch_port}?username=${ch_username}&password=${ch_password}&database=${ch_database}&secure=true\"}}"
    fi
}

# Функция деплоя лоадеров
deploy_loaders() {
    echo "📥 Deploying loaders..."

    local loader_files=(
        "$K8S_DIR/deployments/kafka-loader-deployment.yaml"
        "$K8S_DIR/services/kafka-loader-service.yaml"
        "$K8S_DIR/deployments/clickhouse-loader-deployment.yaml"
        "$K8S_DIR/services/clickhouse-loader-service.yaml"
    )

    for file in "${loader_files[@]}"; do
        if [ -f "$file" ]; then
            kubectl apply -f "$file"
            echo "✅ Applied: $(basename "$file")"
        else
            echo "⚠️ Loader file not found: $file"
        fi
    done

    echo "⏳ Waiting for loaders to start..."
    kubectl rollout status deployment/kafka-loader -n "$NAMESPACE" --timeout=180s
    kubectl rollout status deployment/clickhouse-loader -n "$NAMESPACE" --timeout=180s
    echo "✅ Loaders deployed"
}

# Функция деплоя сервисов
deploy_services() {
    echo "🚀 Deploying microservices..."

    local services=("bid-engine" "orchestrator" "router" "spp-adapter")

    for service in "${services[@]}"; do
        echo "📦 Deploying $service..."

        local deployment_file="$K8S_DIR/deployments/${service}-deployment.yaml"
        local service_file="$K8S_DIR/services/${service}-service.yaml"
        local deployment_name=${service}-deployment

        if [ "$service" = "spp-adapter" ]; then
            if ! ensure_geoip_secret; then
                echo "❌ Cannot deploy spp-adapter without GeoIP database"
                return 1
            fi
        fi

        if [ -f "$deployment_file" ]; then
            kubectl apply -f "$deployment_file"
        else
            echo "❌ Deployment file not found: $deployment_file"
            return 1
        fi

        if [ -f "$service_file" ]; then
            kubectl apply -f "$service_file"
        else
            echo "❌ Service file not found: $service_file"
            return 1
        fi

        kubectl rollout status "deployment/$deployment_name" -n "$NAMESPACE" --timeout=180s
    done

    echo "📊 Services status:"
    kubectl get pods -n "$NAMESPACE"
    echo "✅ Services deployed"
}

ensure_geoip_secret() {
    if kubectl get secret geoip-db -n "$NAMESPACE" >/dev/null 2>&1; then
        echo "✅ GeoIP secret already exists"
        return 0
    fi

    local candidate
    local candidates=(
        "$SCRIPT_DIR/GeoIP2_City.mmdb"
        "$SCRIPT_DIR/GeoLite2-City.mmdb"
        "$ASSETS_DIR/GeoIP2_City.mmdb"
        "$ASSETS_DIR/GeoLite2-City.mmdb"
    )

    for candidate in "${candidates[@]}"; do
        if [ -f "$candidate" ]; then
            echo "🔑 Creating geoip-db secret from $(basename "$candidate")..."
            kubectl create secret generic geoip-db \
                --namespace "$NAMESPACE" \
                --from-file=GeoIP2_City.mmdb="$candidate" \
                --dry-run=client -o yaml | kubectl apply -f -
            echo "✅ GeoIP secret created"
            return 0
        fi
    done

    cat <<'EOF'
⚠️ GeoIP database secret 'geoip-db' is missing.
Положите файл GeoIP2_City.mmdb или GeoLite2-City.mmdb в корень репозитория рядом с deploy.sh (его можно хранить в git).

Также допустимо использовать каталог deploy/assets, если нужно отделить артефакты.

Создать секрет вручную можно командой:
  kubectl create secret generic geoip-db -n exchange \
    --from-file=GeoIP2_City.mmdb=/путь/к/GeoIP2_City.mmdb
EOF

    return 1
}

# Функция деплоя внешнего шлюза
deploy_gateway() {
    echo "🌉 Deploying external gateway..."

    local config_file="$K8S_DIR/configs/gateway-config.yaml"
    local deployment_file="$K8S_DIR/deployments/gateway-deployment.yaml"
    local service_file="$K8S_DIR/services/gateway-service.yaml"

    for file in "$config_file" "$deployment_file" "$service_file"; do
        if [ ! -f "$file" ]; then
            echo "❌ Gateway file not found: $file"
            return 1
        fi
    done

    kubectl apply -f "$config_file"
    kubectl apply -f "$deployment_file"
    kubectl apply -f "$service_file"

    kubectl rollout status deployment/gateway-deployment -n "$NAMESPACE" --timeout=180s
    echo "✅ External gateway is ready"
}

# Функция деплоя ingress
deploy_ingress() {
    echo "🌐 Deploying ingress..."
    if [ -d "$K8S_DIR/ingress" ]; then
        kubectl apply -f "$K8S_DIR/ingress/"
        echo "✅ Ingress deployed"
    else
        echo "❌ Ingress directory not found: $K8S_DIR/ingress/"
        return 1
    fi
}

# Функция проверки статуса
check_status() {
    echo "📊 Deployment status:"
    echo ""
    echo "=== Namespaces ==="
    kubectl get namespaces | grep -E "(NAME|$NAMESPACE)"
    echo ""
    echo "=== Pods ==="
    kubectl get pods -n "$NAMESPACE"
    echo ""
    echo "=== Services ==="
    kubectl get services -n "$NAMESPACE"
    echo ""
    echo "=== Deployments ==="
    kubectl get deployments -n "$NAMESPACE"
    echo ""
    echo "=== StatefulSets ==="
    kubectl get statefulsets -n "$NAMESPACE" 2>/dev/null || echo "No statefulsets found"
    echo ""
    echo "=== Ingress ==="
    kubectl get ingress -n "$NAMESPACE" 2>/dev/null || echo "No ingress found"
}

# Функция показа логов
show_logs() {
    local service="${1:-}"
    local services=("bid-engine" "orchestrator" "router" "spp-adapter" "redis" "kafka" "kafka-loader" "clickhouse-loader" "gateway")
    
    if [ -z "$service" ]; then
        echo "Available services: ${services[*]}"
        echo "Usage: $0 logs [service-name]"
        return 1
    fi
    
    echo "📋 Logs for $service:"
    if [ "$service" = "kafka" ]; then
        kubectl logs -l app=kafka -n "$NAMESPACE" --tail=50 --prefix=true
    elif [ "$service" = "gateway" ]; then
        kubectl logs -l app=gateway -n "$NAMESPACE" --tail=50
    else
        kubectl logs -l app="$service" -n "$NAMESPACE" --tail=50
    fi
}

# Функция тестирования endpoints
test_endpoints() {
    echo "🧪 Testing endpoints..."

    if ! command -v kubectl >/dev/null 2>&1; then
        echo "❌ kubectl не найден. Установите kubectl, чтобы использовать тесты."
        return 1
    fi

    if ! kubectl get nodes >/dev/null 2>&1; then
        echo "❌ Не удалось подключиться к кластеру Kubernetes"
        return 1
    fi

    local node_ip
    node_ip=$(kubectl get nodes -o wide | grep 'Ready' | head -1 | awk '{print $6}')
    if [ -z "$node_ip" ]; then
        node_ip="127.0.0.1"
    fi

    echo "Node IP: $node_ip"

    local gateway_host
    gateway_host=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || true)
    if [ -z "$gateway_host" ] || [ "$gateway_host" = "<no value>" ]; then
        gateway_host=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' 2>/dev/null || true)
    fi
    if [ -z "$gateway_host" ] || [ "$gateway_host" = "<no value>" ]; then
        gateway_host=$node_ip
    fi

    get_gateway_port() {
        local port_name="$1"
        local default_value="$2"
        local jsonpath="{.spec.ports[?(@.name==\\\"$port_name\\\")].port}"
        local value
        value=$(kubectl get svc gateway-service -n "$NAMESPACE" -o jsonpath="$jsonpath" 2>/dev/null || true)
        if [ -z "$value" ]; then
            value="$default_value"
        fi
        echo "$value"
    }

    local http_port=$(get_gateway_port http 80)
    local bid_port=$(get_gateway_port bid-engine 8080)
    local orchestrator_port=$(get_gateway_port orchestrator 8081)
    local router_port=$(get_gateway_port router 8082)
    local spp_port=$(get_gateway_port spp-adapter 8083)

    echo "Gateway host: $gateway_host"
    echo ""

    local endpoints=(
        "gateway:http://$gateway_host:$http_port/healthz"
        "bid-engine:http://$gateway_host:$bid_port/health"
        "orchestrator:http://$gateway_host:$orchestrator_port/health"
        "router:http://$gateway_host:$router_port/health"
        "spp-adapter:http://$gateway_host:$spp_port/health"
        "gateway-router:http://$gateway_host:$http_port/router/health"
    )

    for endpoint in "${endpoints[@]}"; do
        local name=$(echo "$endpoint" | cut -d: -f1)
        local url=$(echo "$endpoint" | cut -d: -f2-)

        echo "Testing $name ($url)..."
        if curl -s --connect-timeout 5 "$url" >/dev/null; then
            echo "✅ $name is accessible"
        else
            echo "❌ $name is not accessible"
        fi
    done

    unset -f get_gateway_port
}

# Основная функция деплоя
deploy_all() {
    echo "🚀 Starting full deployment..."
    
    auto_setup_before_deploy

    if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        echo "📦 Creating namespace $NAMESPACE..."
        kubectl create namespace "$NAMESPACE"
    fi
    
    # Деплоим компоненты в правильном порядке
    deploy_configs
    deploy_redis
    deploy_kafka
    
    # Спрашиваем про ClickHouse Cloud
    echo ""
    read -p "Configure ClickHouse Cloud? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        setup_clickhouse_cloud
    else
        echo "ℹ️ Skipping ClickHouse Cloud configuration"
    fi
    
    deploy_loaders
    deploy_services
    deploy_gateway
    deploy_ingress
    
    echo "✅ Full deployment completed!"
    echo ""
    check_status
    echo ""
    echo "🎉 Deployment ready! Use './deploy.sh test' to test endpoints"
}

# Обработка команд
case "${1:-all}" in
    "all")
        deploy_all
        ;;
    "configs")
        auto_setup_before_deploy
        deploy_configs
        ;;
    "redis")
        auto_setup_before_deploy
        deploy_redis
        ;;
    "kafka")
        auto_setup_before_deploy
        deploy_kafka
        ;;
    "clickhouse")
        auto_setup_before_deploy
        setup_clickhouse_cloud
        ;;
    "loaders")
        auto_setup_before_deploy
        deploy_loaders
        ;;
    "services")
        auto_setup_before_deploy
        deploy_services
        ;;
    "gateway")
        auto_setup_before_deploy
        deploy_gateway
        ;;
    "ingress")
        auto_setup_before_deploy
        deploy_ingress
        ;;
    "status")
        check_status
        ;;
    "logs")
        show_logs "$2"
        ;;
    "test")
        test_endpoints
        ;;
    "clean")
        clean_resources
        ;;
    "destroy")
        destroy_namespace
        ;;
    "help"|"-h"|"--help")
        usage
        ;;
    *)
        echo "Unknown option: $1"
        usage
        exit 1
        ;;
esac
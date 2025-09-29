#!/bin/bash
set -e

REGISTRY="localhost:5000/exchange"
TAG="latest"
SERVICES=("bid-engine" "orchestrator" "router" "spp-adapter")
REGISTRY_CONTAINER="rtb-registry"

echo "=== Building Docker Images ==="

usage() {
    echo "Usage: $0 [all|list|image-name|clean|clean-all|registry-start|registry-stop|registry-status|registry-clean|push-local]"
    echo "  all           - Build all images (default)"
    echo "  list          - List available services"
    echo "  image-name    - Build specific image (e.g., bid_engine)"
    echo "  clean         - Remove all RTB images from local Docker and registry"
    echo "  clean-all     - Remove ALL unused Docker images, containers, registry + k3s cache"
    echo "  registry-start- Start local Docker registry"
    echo "  registry-stop - Stop local Docker registry"
    echo "  registry-status - Check registry status"
    echo "  registry-clean - Clean only registry images"
    echo "  push-local    - Build and push to local registry"
}

# Функция для очистки кеша k3s/containerd
clean_k3s_cache() {
    echo "🧹 Cleaning k3s/containerd cache..."
    
    # 1. Останавливаем k3s
    if systemctl is-active k3s >/dev/null 2>&1; then
        echo "🛑 Stopping k3s..."
        sudo systemctl stop k3s
        sleep 5
    fi
    
    # 2. Запускаем k3s
    echo "🚀 Starting k3s..."
    sudo systemctl start k3s
    
    # Ждем запуска k3s
    echo "⏳ Waiting for k3s to start..."
    for i in {1..30}; do
        if kubectl get nodes >/dev/null 2>&1; then
            echo "✅ k3s is running"
            break
        fi
        sleep 2
    done
    
    # 3. ТЕПЕРЬ удаляем образы через ctr (когда k3s работает)
    echo "🗑️ Removing images via ctr..."
    if command -v ctr >/dev/null 2>&1; then
        # Удаляем все образы с меткой exchange
        sudo ctr -n k8s.io images ls | grep exchange | awk '{print $1}' | while read image; do
            echo "Removing: $image"
            sudo ctr -n k8s.io images rm "$image" 2>/dev/null || true
        done
    fi
    
    echo "✅ k3s cache cleanup completed"
    return 0
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

# Функция проверки registry
check_registry() {
    docker ps | grep -q $REGISTRY_CONTAINER
}

# Функция запуска registry
start_registry() {
    if check_registry; then
        echo "✅ Registry is already running"
        return 0
    fi
    
    echo "🚀 Starting local registry..."
    docker run -d \
        --name $REGISTRY_CONTAINER \
        --restart=always \
        -p 5000:5000 \
        -v $(pwd)/registry-data:/var/lib/registry \
        registry:2
    
    # Ждем пока registry запустится
    for i in {1..10}; do
        if curl -s http://localhost:5000/v2/_catalog >/dev/null; then
            echo "✅ Registry started successfully"
            return 0
        fi
        sleep 1
    done
    echo "❌ Failed to start registry"
    return 1
}

# Функция остановки registry
stop_registry() {
    if check_registry; then
        echo "🛑 Stopping registry..."
        docker stop $REGISTRY_CONTAINER
        docker rm $REGISTRY_CONTAINER
        echo "✅ Registry stopped"
    else
        echo "ℹ️ Registry is not running"
    fi
}

# Функция проверки статуса registry
registry_status() {
    if check_registry; then
        echo "✅ Registry is running"
        echo "📦 Registry contents:"
        curl -s http://localhost:5000/v2/_catalog | python3 -m json.tool 2>/dev/null || curl -s http://localhost:5000/v2/_catalog
    else
        echo "❌ Registry is not running"
    fi
}

# Функция сборки одного образа
build_image() {
    local service=$1
    local image_name="${REGISTRY}/${service}:${TAG}"
    local dockerfile_path="deploy/docker/${service}.dockerfile"
    
    echo "🏗️ Building ${service}..."
    echo "📁 Using Dockerfile: ${dockerfile_path}"
    
    # Проверяем существование Dockerfile
    if [ ! -f "$dockerfile_path" ]; then
        echo "❌ Dockerfile not found: $dockerfile_path"
        echo "Available Dockerfiles:"
        ls deploy/docker/*.dockerfile 2>/dev/null || echo "No Dockerfiles found in deploy/docker/"
        return 1
    fi
    
    # Собираем образ с правильным контекстом (корень проекта)
    docker build -t $image_name -f $dockerfile_path .
    echo "✅ Built ${image_name}"
}

# Функция пуша одного образа в локальный registry
push_to_local() {
    local service=$1
    local image_name="${REGISTRY}/${service}:${TAG}"
    
    echo "📤 Pushing ${service} to local registry..."
    docker push $image_name
    echo "✅ Pushed ${image_name}"
}

# Функция сборки всех образов
build_all() {
    echo "🏗️ Building all services..."
    for service in "${SERVICES[@]}"; do
        build_image $service
    done
    echo "✅ All images built"
}

# Функция пуша всех образов в локальный registry
push_all_local() {
    echo "📤 Pushing all images to local registry..."
    for service in "${SERVICES[@]}"; do
        push_to_local $service
    done
    echo "✅ All images pushed to registry"
}

# Функция очистки registry
clean_registry() {
    echo "🧹 Cleaning registry..."
    if check_registry; then
        docker exec $REGISTRY_CONTAINER registry garbage-collect /etc/docker/registry/config.yml --delete-untagged=true
        echo "✅ Registry cleaned"
    else
        echo "ℹ️ Registry is not running"
    fi
}

# Функция очистки образов
clean_images() {
    echo "🧹 Removing RTB images..."
    for service in "${SERVICES[@]}"; do
        local image_name="${REGISTRY}/${service}:${TAG}"
        if docker image inspect $image_name >/dev/null 2>&1; then
            docker rmi $image_name
            echo "✅ Removed ${image_name}"
        fi
    done
}

# Функция полной очистки Docker и k3s кеша
clean_all_docker() {
    echo "🧹 Cleaning all Docker resources..."
    
    # Останавливаем и удаляем registry
    stop_registry
    
    # Удаляем все контейнеры
    if [ "$(docker ps -aq)" ]; then
        docker stop $(docker ps -aq)
        docker rm $(docker ps -aq)
        echo "✅ All containers removed"
    fi
    
    # Удаляем все образы
    if [ "$(docker images -q)" ]; then
        docker rmi -f $(docker images -q)
        echo "✅ All images removed"
    fi
    
    # Очищаем volumes и network
    docker system prune -a --volumes -f
    echo "✅ Docker system cleaned"
    
    # Очищаем кеш k3s/containerd
    echo "🧹 Cleaning k3s cache..."
    if clean_k3s_cache; then
        echo "✅ k3s cache cleaned successfully"
    else
        echo "❌ Failed to clean k3s cache"
        return 1
    fi
    
    echo "✅ Full cleanup completed - both Docker and k3s cache are clean"
}

# Автоматическая настройка перед push-local
auto_setup_before_push() {
    echo "🔧 Auto-setting up environment for local registry..."
    
    # Настраиваем k3s если он установлен
    setup_k3s_registry
    
    # Запускаем registry если не запущен
    if ! check_registry; then
        start_registry
    fi
    
    echo "✅ Environment setup completed"
}

# Основная функция для push-local
build_and_push_local() {
    echo "🏗️ Building and pushing to local registry..."
    
    # Автоматическая настройка
    auto_setup_before_push
    
    # Собираем образы
    build_all
    
    # Пушим в локальный registry
    push_all_local
    
    echo "✅ All images built and pushed to local registry"
    registry_status
}

# Обработка команд
case "${1:-all}" in
    "all")
        build_all
        echo ""
        read -p "Push images to local registry? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            build_and_push_local
        fi
        ;;
    "push-local")
        build_and_push_local
        ;;
    "list")
        echo "Available services:"
        for service in "${SERVICES[@]}"; do
            echo "  - $service"
        done
        ;;
    "clean")
        clean_images
        ;;
    "clean-all")
        clean_all_docker
        ;;
    "registry-start")
        start_registry
        ;;
    "registry-stop")
        stop_registry
        ;;
    "registry-status")
        registry_status
        ;;
    "registry-clean")
        clean_registry
        ;;
    "help"|"-h"|"--help")
        usage
        ;;
    *)
        # Assume it's a service name
        if [[ " ${SERVICES[@]} " =~ " $1 " ]]; then
            build_image $1
            read -p "Push image to local registry? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                auto_setup_before_push
                push_to_local $1
            fi
        else
            echo "❌ Unknown service: $1"
            usage
            exit 1
        fi
        ;;
esac
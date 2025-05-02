# Load Balancer

Load Balancer — это высокопроизводительный балансировщик нагрузки, который поддерживает различные стратегии распределения запросов, такие как **Round Robin** и **Least Connections**, а также включает встроенный механизм ограничения запросов (Rate Limiter) и проверки состояния (Health Checks) для бэкендов.


### Требования
- Go версии 1.18 или выше.
- Docker (опционально, для запуска через контейнеры).
### Установка
1. Клонируйте репозиторий:
    ```bash
    git clone https://github.com/your-repo/load_balancer.git
    cd load_balancer
    ```
2. Соберите проект:
    ```bash
    go build -o load_balancer
    ```

### Запуск
1. Запустите бинарный файл:
    ```bash
    ./load_balancer -config=config.yaml
    ```
2. Для запуска через Docker:
    ```bash
    docker build -t load_balancer .
    docker run -d -p 8080:8080 load_balancer
    ```

## Пример конфигурации

Пример файла `config.yaml`:
```yaml
backends:
  - url: http://backend1.example.com
     weight: 1
  - url: http://backend2.example.com
     weight: 2

strategy: round_robin

rate_limiter:
  enabled: true
  requests_per_second: 100

health_checks:
  enabled: true
  interval: 10s
```

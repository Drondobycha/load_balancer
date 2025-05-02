# Load Balancer

Load Balancer — это высокопроизводительный балансировщик нагрузки, поддерживающий различные стратегии распределения запросов, такие как **Round Robin** и **Least Connections**. Он также включает встроенные механизмы ограничения запросов (Rate Limiter) и проверки состояния (Health Checks) для бэкендов.


### Установка
1. Клонируйте репозиторий:
    ```bash
    git clone https://github.com/Drondobycha/load_balancer/tree/develop
    cd load_balancer
    ```
2. Соберите проект:
    ```bash
    go build cmd/main.go
    ```

### Запуск
1. Запустите бинарный файл:
    ```bash
    ./main -config=config.yaml
    ```
2. Для запуска через Docker:
    ```bash
    docker compose up --build
    ```
3. Для провперки работоспособности
    ```bash
    ab -n 5000 -c 1000 http://localhost:8080/
    ```




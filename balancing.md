# Балансировка и отказоустойчивость

1. Были подняты реплики базы
2. Изменен docker-compose.yml (добавлены реплики, haproxy, nginx)
2. Соединение со слейвами реализовано через HAProxy
```
global
   log stdout format raw local0
   maxconn 100

defaults
log global
mode tcp
retries 2
timeout connect 3000ms
timeout server 5000ms
timeout client 5000ms

frontend main_front
bind *:5000
default_backend db_back

backend db_back
mode tcp
balance roundrobin
option tcp-check
tcp-check connect
server pg1 pg:5432 maxconn 100 check
server pg2 pg_replica:5432 maxconn 100 check
server pg3 pg_replica2:5432 maxconn 100 check
```

3. Создана копия приложения app (app2) и добавлена конфигурация nginx
```
worker_processes 1;

events {
    worker_connections 1024;
}

http {
    upstream app_servers {
        least_conn;
        server app:8080;
        server app2:8080;
    }

    server {
        listen 80;

        location / {
            proxy_pass http://app_servers;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```
4. Проверка балансировки.

Проводил через старый скрипт k6s по поиску юзера. Под нагрузкой были отключены pg_replica и app2 в результате чего это не заафектило работу приложения. Приложение осталось работоспособным и отвечало на запросы.


# Копия базы DEEIX

База (SQLite) и загруженные файлы лежат в томе Docker `deeix-chat-app-data`
(`/var/lib/docker/volumes/deeix-chat-app-data/_data`). Копия делается при
остановленном чате (на 10–30 секунд), чтобы база не записывалась во время копирования.

## Сделать копию (терминал 1Panel)
```
clear && mkdir -p /opt/deeix-backup && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && tar -czf /opt/deeix-backup/deeix-data-$(date +%F-%H%M).tar.gz -C /var/lib/docker/volumes/deeix-chat-app-data/_data . ; $C up -d app && sleep 10 && curl -s http://127.0.0.1:8080/healthz && echo && ls -lh /opt/deeix-backup
```
В конце: `{"status":"ok",...}` и файл `deeix-data-ДАТА-ВРЕМЯ.tar.gz`.

Копия на том же сервере не спасёт, если пропадёт сам VPS: скачайте файл
через 1Panel → «Файлы» → `/opt/deeix-backup` → «Скачать» к себе.

## Восстановить из копии
Подставьте имя файла вместо `ФАЙЛ`. Текущая база перед заменой сохраняется в `_data-pered-vosstanovleniem`.
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && V=/var/lib/docker/volumes/deeix-chat-app-data && $C stop app && rm -rf $V/_data-pered-vosstanovleniem && mv $V/_data $V/_data-pered-vosstanovleniem && mkdir $V/_data && tar -xzf /opt/deeix-backup/ФАЙЛ -C $V/_data && $C up -d app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

## Автоматическая копия раз в сутки
Скрипт `/opt/deeix-backup/backup.sh` (ставится командой из чата 2026-10-22), запуск через
`/etc/cron.d/deeix-backup` каждый день в 04:00 по времени сервера. Хранятся последние 7 копий,
журнал в `/opt/deeix-backup/backup.log` (строка FAILED значит, что копия не сделана).

Проверить: `tail /opt/deeix-backup/backup.log; ls -lh /opt/deeix-backup`
Отключить: `rm /etc/cron.d/deeix-backup`

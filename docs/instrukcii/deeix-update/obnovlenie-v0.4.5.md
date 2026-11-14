# Обновление DEEIX на VPS до v0.4.5 (наша сборка)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/39498805066 (ветка `update/v0.4.5`, PR #5).

## Шаг 1. Скачать и положить файлы
1. Откройте ссылку на сборку выше, внизу в разделе Artifacts скачайте `deeix-ru-out` и `deeix-ru-backend` (придут файлы `deeix-ru-out.zip` и `deeix-ru-backend.zip`).
2. В 1Panel → «Файлы» загрузите оба zip в папку `/opt/deeix-ru/`.

## Шаг 2. Одна команда в терминале 1Panel
Делает копии старой версии и базы (`...-0.4.4`), подтягивает новый образ, ставит нашу сборку и показывает версию.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C pull app && $C stop app && cp -a /var/lib/docker/volumes/deeix-chat-app-data/_data /opt/deeix-ru/data-0.4.4 && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-0.4.4 && mv /opt/deeix-ru/out /opt/deeix-ru/out-0.4.4 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok","version":"0.4.5"...}` (или похожее с 0.4.5). Если пишет `unzip: command not found`, сначала выполните `apt-get install -y unzip` и повторите.

## Шаг 3. Проверить на сайте
Обновите страницу (Ctrl+F5): русский язык, иконка «D», отправка PDF-скана, DeepSeek с pravo.

## Если что-то сломалось — откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /var/lib/docker/volumes/deeix-chat-app-data/_data && cp -a /opt/deeix-ru/data-0.4.4 /var/lib/docker/volumes/deeix-chat-app-data/_data && cp -a /opt/deeix-ru/deeix-chat-0.4.4 /opt/deeix-ru/deeix-chat && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-0.4.4 /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

## Уборка (через пару дней, когда всё работает)
```
clear && rm -rf /opt/deeix-ru/data-0.4.4 /opt/deeix-ru/out-0.4.4 /opt/deeix-ru/deeix-chat-0.4.4 /opt/deeix-ru/deeix-ru-out.zip /opt/deeix-ru/deeix-ru-backend.zip
```

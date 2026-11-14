# Установка патча «HTML-вёрстка в ответе» (PR #7)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/39959444086 (ветка `feat/inline-html`).
В ней все прежние патчи (включая таблицы) плюс новый. Меняется только сайт (`out`), бэкенд и база не затрагиваются.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-html`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-html && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-html && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и откройте тот же диалог «Идеи для онлайн-сервиса и дохода»: вместо кода «Три точки входа» должны быть карточки прямо в ответе. Старые ответы тоже перерисуются, переспрашивать модель не нужно.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-html /opt/deeix-ru/out && $C up -d --force-recreate app
```

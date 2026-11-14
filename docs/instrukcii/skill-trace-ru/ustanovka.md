# Установка: перевод строки про навыки в ходе обработки (PR #14)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/40659070898 (ветка `fix/skill-trace-ru`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс перевод.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-perevodom`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-perevodom && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-perevodom && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте любой ответ со скиллом: под «Обработка завершена» вместо «已提供 1 个 Skill 上下文» будет «Подключён 1 навык». Если раскрыть, будет «Контекст навыков: Для этого ответа подключены: …». Старые ответы тоже переводятся, база не меняется.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-perevodom /opt/deeix-ru/out && $C up -d --force-recreate app
```

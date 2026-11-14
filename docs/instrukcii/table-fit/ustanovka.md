# Установка патча для широких таблиц (PR #6)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/39790501102 (ветка `feat/table-fit`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (если там лежит старый `deeix-ru-out.zip`, замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-tablicami`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-tablicami && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-tablicami && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и откройте ответ с широкой таблицей: она должна помещаться в ширину чата, текст в ячейках переноситься.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-tablicami /opt/deeix-ru/out && $C up -d --force-recreate app
```

# Таблицы-реестры: перенос текста и «Открыть во всю ширину» (PR #24)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42496055924 (ветка `feat/data-table-fit`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются.

## Шаг 1. Скачать
1. Откройте ссылку на сборку (дождитесь зелёной галочки), внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (если там лежит старый `deeix-ru-out.zip`, замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-reestrom`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-reestrom && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-reestrom && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и откройте ответ с реестром:
- длинный текст в ячейках переносится на несколько строк, таблица в ширину чата;
- под таблицей (на компьютере при наведении) кнопка «Открыть во всю ширину» — таблица открывается на весь экран, закрыть крестиком или Esc.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-reestrom /opt/deeix-ru/out && $C up -d --force-recreate app
```

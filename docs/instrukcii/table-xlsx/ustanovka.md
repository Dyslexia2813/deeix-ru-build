# Скачать таблицу в Excel (PR #30)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/30
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/44762969891
Меняется только сайт (`out`), бэкенд и база не затрагиваются.

## Шаг 1. Скачать
1. Откройте ссылку на сборку (дождитесь зелёной галочки), внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (если там лежит старый `deeix-ru-out.zip`, замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-excel`, ставит новый, перезапускает контейнер и проверяет, что новый код на месте.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-excel && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-excel && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && echo "новый код: $(grep -rl 'spreadsheetml.sheet.main+xml' /opt/deeix-ru/out/_next | wc -l)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код: 1` (или больше) и в конце `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и откройте ответ с таблицей:
- у обычной таблицы значок скачивания справа вверху: меню «Скачать Excel / CSV / Markdown»;
- у реестра (со строкой поиска): меню «Скачать Excel / Скачать CSV»;
- «Скачать Excel» сохраняет .xlsx: строка заголовков жирная, закреплена и с фильтром, суммы считаются числами, ИНН остаётся как есть.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-excel /opt/deeix-ru/out && $C up -d --force-recreate app
```

## Уборка (когда всё проверено)
```
clear && rm -rf /opt/deeix-ru/out-pered-excel /opt/deeix-ru/deeix-ru-out.zip && echo готово
```

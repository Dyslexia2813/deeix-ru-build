# Имя файла Excel по заголовку таблицы (PR #31)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/31
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/44775203371
Меняется только сайт (`out`), как в прошлый раз.

## Шаг 1. Скачать
1. Откройте ссылку на сборку (дождитесь зелёной галочки), внизу в разделе Artifacts скачайте только `deeix-ru-out`.
2. В 1Panel → «Файлы» загрузите `deeix-ru-out.zip` в `/opt/deeix-ru/` (старый замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-imya`, ставит новый, проверяет новый код и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-imya && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-imya && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && echo "новый код: $(grep -rlF 'H[1-6]' /opt/deeix-ru/out/_next | wc -l)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код: 1` (или больше) и в конце `{"status":"ok",...}`.

## Шаг 3. Проверить
Ctrl+F5, затем «Скачать Excel» у таблицы, над которой есть заголовок или подпись с двоеточием. Файл называется так же, например «Задачи юристов на декабрь.xlsx». Если над таблицей ничего такого нет, файл, как и раньше, называется «Таблица.xlsx».

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-imya /opt/deeix-ru/out && $C up -d --force-recreate app
```

## Уборка (когда всё проверено)
Удаляет и копию от прошлой установки.
```
clear && rm -rf /opt/deeix-ru/out-pered-imya /opt/deeix-ru/out-pered-excel /opt/deeix-ru/deeix-ru-out.zip && echo готово
```

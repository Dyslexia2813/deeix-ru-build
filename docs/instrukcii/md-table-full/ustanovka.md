# Установка: заголовки обычных таблиц и «Открыть во всю ширину» (PR #35)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/35
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/45622822425 (ветка `claude/project-thread-l5si3q`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс эту правку.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-35`, ставит новый, проверяет, что в нём новый код, и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-35 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-35 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && echo "новый код: $(grep -rl table-fullscreen-action /opt/deeix-ru/out/_next | wc -l)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть «новый код: 1» (или больше) и в конце `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте чат с той таблицей («Что формулировка покрывает, а что нет»).

- Длинные заголовки столбцов переносятся на 2–3 строки и больше не заходят друг на друга.
- На компьютере наведите мышь на таблицу: справа под ней появится «Открыть во всю ширину». На телефоне эта надпись видна всегда.
- Нажмите её: таблица откроется на весь экран, сверху её название, кнопка скачивания и крестик. На телефоне таблицу там можно листать вбок.

Примеры: `pk.png`, `pk-polnyj-ekran.png`, `telefon.png`, `telefon-polnyj-ekran.png` в этой папке.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-35 /opt/deeix-ru/out && $C up -d --force-recreate app
```

## Уборка (когда всё хорошо)
```
clear && rm -rf /opt/deeix-ru/out-pered-35 && echo готово
```

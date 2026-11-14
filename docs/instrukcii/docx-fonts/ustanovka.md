# Установка: экспорт в Word со шрифтами с экрана (PR #9)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/40180384706 (ветка `feat/docx-screen-fonts`).
В ней все прежние патчи (таблицы, HTML-вёрстка, кнопки / и @) плюс новый экспорт. Меняется только сайт (`out`), бэкенд и база не затрагиваются.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-docx`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-docx && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-docx && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5), под любым ответом нажмите «Экспорт в Word (DOCX)». Лучше взять ответ с HTML-карточками или цветным текстом.
В Word должны совпасть с экраном шрифт, размер, цвета, выделение, карточки (рамка и заливка), фон ячеек таблиц.
Экспортируйте в светлой теме: в тёмной цвета намеренно не переносятся, чтобы не было белого текста на белой странице.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-docx /opt/deeix-ru/out && $C up -d --force-recreate app
```

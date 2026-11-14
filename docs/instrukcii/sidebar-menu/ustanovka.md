# Установка: компактное верхнее меню боковой панели (PR #21)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42401926404 (ветка `sidebar-menu`).
Меняется только сайт (`out`), бэкенд, OCR и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main, включая ширину панели из #20), плюс компактное меню.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените; если браузер добавил к имени « (1)», уберите это).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-menu`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-menu && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-menu && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5). Вверху боковой панели остался только «Новый чат», справа от него стрелка ⌄.

- Нажмите стрелку: раскроются «Поиск», «Недавние», «Файлы», «Базы знаний», «Библиотека». Ещё раз — свернутся.
- Выбор запоминается в этом браузере (после перезагрузки останется как было).
- На телефоне то же самое, по нажатию.
- Ctrl+K (поиск) и Ctrl+Shift+O (новый чат) работают всегда.
- В свёрнутой до иконок панели (◫ или Ctrl+B) все иконки видны, как раньше.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-menu /opt/deeix-ru/out && $C up -d --force-recreate app
```

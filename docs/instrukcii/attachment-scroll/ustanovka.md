# Установка: прокрутка вложений в поле ввода (PR #22)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/43482039949 (ветка `claude/attachment-scroll`).
Меняется только сайт (`out`), бэкенд, OCR и база не затрагиваются. Внутри всё, что уже стоит на сервере (main до #26 включительно), плюс прокрутка вложений.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените; если браузер добавил к имени « (1)», уберите это).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-vlozheniyami`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-vlozheniyami && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-vlozheniyami && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и прикрепите 5–6 файлов: колесо мыши листает ряд вбок, внизу тонкая полоса, справа «N файлов».

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-vlozheniyami /opt/deeix-ru/out && $C up -d --force-recreate app
```

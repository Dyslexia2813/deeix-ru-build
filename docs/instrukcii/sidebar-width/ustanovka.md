# Установка: ширина боковой панели мышкой (PR #20)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42305192737 (ветка `sidebar-width`).
Меняется только сайт (`out`), бэкенд, OCR и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main, включая своё OCR из #19), плюс перетаскиваемый край панели.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените; если браузер добавил к имени « (1)», уберите это).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-shirinoy`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-shirinoy && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-shirinoy && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5). На компьютере при открытой боковой панели наведите мышь на её правый край: курсор станет ↔, край подсветится. Зажмите и тяните вправо или влево.

- Ширина от 220 до 560 px, чату всегда остаётся не меньше 360 px.
- После перезагрузки ширина сохраняется (запоминается в этом браузере).
- Двойной щелчок по краю возвращает обычную ширину.
- Свёрнутая панель (◫ или Ctrl+B) и телефон работают как раньше.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-shirinoy /opt/deeix-ru/out && $C up -d --force-recreate app
```

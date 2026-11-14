# Установка: поле ввода внизу на пустом экране (PR #13)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/40348888795 (ветка `feat/composer-bottom`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс новое расположение поля.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-polem`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-polem && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-polem && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). На пустом экране нового чата приветствие посередине, поле ввода внизу. Наберите «/» (или «+» → «/ Скиллы и промпты»): меню скиллов раскроется вверх, во всю высоту. После отправки первого сообщения поле остаётся на месте.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-polem /opt/deeix-ru/out && $C up -d --force-recreate app
```

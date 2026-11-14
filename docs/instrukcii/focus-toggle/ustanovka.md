# Установка: плавающая кнопка «скрыть заголовок и поле ввода» (PR #16)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/41306400052 (ветка `feat/focus-toggle`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс новую кнопку. PR #15 в неё не входит.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-knopkoy`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-knopkoy && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-knopkoy && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте любой чат с сообщениями. Внизу над полем ввода, по центру будет круглая кнопка со стрелкой вверх-вниз (↕); когда вы прокрутите чат вверх, рядом с ней появится «Прокрутить в конец», и обе встанут парой по центру. Нажмите: название чата вверху и поле ввода исчезнут, кнопка останется внизу. Нажмите ещё раз, и всё вернётся. Примеры: `pk-do.png`, `pk-posle.png`, `telefon-do.png`, `telefon-posle.png` в этой папке.

После перезагрузки или перехода в другой чат заголовок и поле снова видны.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-knopkoy /opt/deeix-ru/out && $C up -d --force-recreate app
```

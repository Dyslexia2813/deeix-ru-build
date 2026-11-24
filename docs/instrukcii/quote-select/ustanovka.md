# Установка: «Цитировать» выделенный текст ответа (PR #34)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/34
Сборка: RUN_LINK (ветка `claude/project-thread-q853j0`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс цитирование.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-citatoy`, ставит новый, проверяет, что в нём новый код, и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-citatoy && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-citatoy && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && echo "новый код: $(grep -rl data-chat-quote-button /opt/deeix-ru/out/_next | wc -l)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть «новый код: 1» (или больше) и в конце `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте чат с ответом модели.

- На компьютере выделите мышкой кусок ответа: над ним появится кнопка «Цитировать». Нажмите — в поле ввода встанет `> выделенный текст`, курсор строкой ниже. Допишите вопрос и отправьте. В чате цитата будет с серой полосой слева.
- На телефоне выделите текст пальцем (долгое нажатие): кнопка «Цитировать» появится под выделением, системное меню «Копировать» остаётся сверху.
- Если в поле уже что-то написано, цитата добавится после этого текста.

Примеры: `pk-vydelenie.png`, `pk-pole.png`, `telefon-vydelenie.png`, `telefon-pole.png` в этой папке.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-citatoy /opt/deeix-ru/out && $C up -d --force-recreate app
```

## Уборка (когда всё хорошо)
```
clear && rm -rf /opt/deeix-ru/out-pered-citatoy && echo готово
```

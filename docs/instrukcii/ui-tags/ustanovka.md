# Установка: компоненты вместо HTML-тегов в ответе (PR #33)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/33
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/45266207901 (артефакты `deeix-ru-out` и `deeix-ru-backend`).

Меняются **сайт (`out`)** и **бэкенд (`deeix-chat`)**. База и настройки не меняются. В сборке всё, что уже стоит на сервере (до PR #32 включительно), плюс эта правка.

## Шаг 1. Скачать
1. Откройте https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/45266207901 и внизу в разделе Artifacts скачайте `deeix-ru-out` и `deeix-ru-backend`.
2. В 1Panel → «Файлы» удалите старые zip с такими именами в `/opt/deeix-ru/` и загрузите новые. Если браузер добавил к имени ` (1)`, переименуйте.

## Шаг 2. Поставить
Сохраняет текущий сайт в `out-pered-33`, бинарник в `deeix-chat-pered-33`, ставит новые и перезапускает DEEIX.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out-pered-33 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-33 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-33 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый сайт: $(grep -rlF '\[object Object\]' /opt/deeix-ru/out | wc -l)" && echo "новый код: $(grep -c 'HTML 属性里放不下' /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый сайт:` и `новый код:` с числами больше 0 и в конце `{"status":"ok",...}`. Если где-то 0, в zip старая сборка: скачайте заново.

## Если первая сборка уже стоит (только сайт)
Во второй сборке (2026-11-24) изменился только сайт: таблица, у которой модель потеряла названия столбцов, теперь рисуется. Скачайте из сборки выше только `deeix-ru-out`, загрузите в `/opt/deeix-ru/` вместо старого и выполните (копия `out-pered-33` с самым первым сайтом остаётся для отката):

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Ctrl+F5 — таблица «Оплата по этапам» должна нарисоваться.

## Шаг 3. Проверить
1. Обновите страницу (Ctrl+F5) и откройте тот чат, где был виден `<decision-tree …>`. Вместо текста должно появиться дерево решений «Проверка самозанятого» с кнопками-ответами.
2. В чате, где был `<stat-grid … items="[object Object]…">`, вместо тега будет небольшая серая карточка «блок не отрисован»: данные этого блока модель потеряла ещё при написании, восстановить их нельзя.
3. Новый ответ с включённой визуальной вёрсткой: модель должна писать карточки и деревья сразу нормальными блоками.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-33 /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat-pered-33 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -rf /opt/deeix-ru/out-pered-33 /opt/deeix-ru/deeix-chat-pered-33
```

# Число веток перед названием чата (PR #26)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/26
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42702146329 (ветка `feat/branch-count`).

Меняются **сайт (`out`) и бэкенд (`deeix-chat`)**, база не меняется. В сборке всё, что уже стоит на сервере (до PR #25 включительно), плюс эта правка.

## Шаг 1. Скачать
1. Откройте ссылку на сборку (дождитесь зелёной галочки). Внизу в разделе Artifacts скачайте **оба** файла: `deeix-ru-out` и `deeix-ru-backend`.
2. В 1Panel → «Файлы» удалите старые `/opt/deeix-ru/deeix-ru-out.zip` и `/opt/deeix-ru/deeix-ru-backend.zip` и загрузите новые. Если браузер добавил к имени ` (1)`, переименуйте.

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-26`, бинарник в `deeix-chat-pered-26`, ставит новые и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out-pered-26 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-26 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-26 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c CountConversationBranches /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код:` с числом больше 0 и в конце `{"status":"ok",...}`. Если `новый код: 0`, в zip старый бинарник: скачайте его заново.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5).
- У чатов, где есть перегенерированные ответы (стрелки 1/2 под ответом), в боковой панели перед названием стоит `(2)`, `(3)` и т. д. У чатов с одной веткой номера нет.
- Перегенерируйте ответ в любом чате: через секунду после появления нового варианта номер в боковой панели увеличится.
- Переименование показывает в поле название без номера, так и должно быть.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-26 /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat-pered-26 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -rf /opt/deeix-ru/out-pered-26 /opt/deeix-ru/deeix-chat-pered-26
```

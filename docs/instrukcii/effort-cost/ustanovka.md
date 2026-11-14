# Уровень рассуждения под ответом и стоимость чата в шапке (PR #25)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/25
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42677575533 (ветка `feat/effort-cost`).

Меняются **сайт (`out`) и бэкенд (`deeix-chat`)**, база не меняется. В сборке всё, что уже стоит на сервере (до PR #24 включительно), плюс эта правка.

## Шаг 1. Скачать
1. Откройте ссылку на сборку (дождитесь зелёной галочки). Внизу в разделе Artifacts скачайте **оба** файла: `deeix-ru-out` и `deeix-ru-backend`.
2. В 1Panel → «Файлы» удалите старые `/opt/deeix-ru/deeix-ru-out.zip` и `/opt/deeix-ru/deeix-ru-backend.zip` и загрузите новые. Если браузер добавил к имени ` (1)`, переименуйте.

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-25`, бинарник в `deeix-chat-pered-25`, ставит новые и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out-pered-25 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-25 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-25 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c SumConversationBilledNanousd /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код: 1` и в конце `{"status":"ok",...}`. Если `новый код: 0`, в zip старый бинарник: скачайте его заново.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5) и откройте чат, где есть платные ответы.
- Под ответом рядом с моделью видно `· 🧠 Высокий` (или другой уровень). У старых ответов и у моделей без рассуждения уровня нет, это нормально.
- В шапке слева от значка «Поделиться» видна сумма, например `₽2.98`. При наведении показывается точная сумма. В неё входят и перегенерированные варианты (стрелки 1/2 под ответом).
- Если суммы нет, проверьте, что в Настройках → Чат включён показ стоимости.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-25 /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat-pered-25 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -rf /opt/deeix-ru/out-pered-25 /opt/deeix-ru/deeix-chat-pered-25
```

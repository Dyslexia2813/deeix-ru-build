# Установка: чтение файлов .doc (PR #17)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/17
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/41604586893 (ветка `fix/doc-text`, 2026-11-03, вторая: плюс .doc, которые на деле RTF/HTML/Word 95). Первую сборку 41602101074 не берите.

Меняется только бэкенд (бинарник `deeix-chat`). Сайт (`out`) и база не меняются. В сборке всё, что уже стоит на сервере (до PR #15 и #16 включительно), плюс чтение `.doc`.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-backend` (придёт `deeix-ru-backend.zip`).
2. В 1Panel → «Файлы» сначала удалите старый `/opt/deeix-ru/deeix-ru-backend.zip`, потом загрузите новый. Новый весит 24,24 МБ (старый от PR #15 — 24,22 МБ). Если браузер сохранил файл как `deeix-ru-backend (1).zip`, переименуйте в `deeix-ru-backend.zip`.

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий бинарник как `deeix-chat-pered-17`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-17 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c ExtractDocBinaryText /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код: 1` и в конце `{"status":"ok",...}`. Если `новый код: 0` — в zip старый бинарник, скачайте заново.

## Шаг 3. Проверить
1. В новом чате прикрепите заново `ПРИКАЗ_КИИ_АКС.doc` (тот же файл: неудачная загрузка обрабатывается заново) и спросите, что в нём написано.
2. Ответ должен прийти без «Файл ещё не готов». Просмотр .doc в окне файла по-прежнему «нельзя просмотреть» — это только предпросмотр, на ответ не влияет.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat-pered-17 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -f /opt/deeix-ru/deeix-chat-pered-17
```

# Установка: перевод китайских строк и исправление статистики (PR #15)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/41344704467 (ветка `claude/translate-server-text-usage-stats`, 2026-11-01).
Меняются и сайт (`out`), и бэкенд (бинарник `deeix-chat`). База не меняется. В сборке всё, что уже стоит на сервере, включая плавающую кнопку из PR #16, плюс новое. Ставится поверх #16, ничего удалять не нужно. Старую сборку run 41250985806 не берите: в ней нет кнопки из #16.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте `deeix-ru-out` и `deeix-ru-backend` (придут `deeix-ru-out.zip` и `deeix-ru-backend.zip`).
2. В 1Panel → «Файлы» загрузите оба в `/opt/deeix-ru/` (старые zip замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущие сайт и бинарник (`out-pered-15`, `deeix-chat-pered-15`), ставит новые и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out-pered-15 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-15 && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-15 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
1. Обновите страницу (Ctrl+F5, в приложении кнопка «Обновить»). Тот ответ, где было «请求未完成», теперь показывает «Запрос не выполнен». Старые ответы тоже переводятся.
2. Админка → Статистика открывается без «Внутренняя ошибка сервера».
3. Админка → Файлы, Апстримы, Модели на русском.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-15 /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat-pered-15 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -rf /opt/deeix-ru/out-pered-15 /opt/deeix-ru/deeix-chat-pered-15
```

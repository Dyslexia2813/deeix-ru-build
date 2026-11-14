# Обновление DEEIX на VPS до v0.4.6 (наша сборка со всеми правками)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/41639683403 (ветка `update/v0.4.6`, PR #18).
Внутри все наши правки: русский язык, иконка «D», PDF-сканы, DeepSeek с инструментами, широкие таблицы,
HTML в ответе, «/» и «@» в меню «+», строка «Скиллы / Инструменты», поле ввода внизу, экспорт в Word,
кнопка ↕, перевод текстов сервера и статистика, чтение .doc.

## Шаг 1. Скачать и положить файлы
1. Откройте ссылку на сборку выше (дождитесь зелёной галочки), внизу в разделе Artifacts скачайте
   `deeix-ru-out` и `deeix-ru-backend` (придут `deeix-ru-out.zip` и `deeix-ru-backend.zip`).
2. В 1Panel → «Файлы» загрузите оба zip в папку `/opt/deeix-ru/` (если там остались старые zip — замените).

## Шаг 2. Одна команда в терминале 1Panel
Сначала делает копию базы (`/opt/deeix-ru/baza-pered-0.4.6.tar.gz`), затем копии старой версии
(`...-old`), подтягивает новый образ, ставит нашу сборку и показывает версию.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C pull app && $C stop app && tar -czf /opt/deeix-ru/baza-pered-0.4.6.tar.gz -C /var/lib/docker/volumes/deeix-chat-app-data/_data . && ls -lh /opt/deeix-ru/baza-pered-0.4.6.tar.gz && rm -rf /opt/deeix-ru/deeix-chat-old /opt/deeix-ru/out-old && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-old && mv /opt/deeix-ru/out /opt/deeix-ru/out-old && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok","version":"0.4.6"...}`.
Если пишет `unzip: command not found`, выполните `apt-get install -y unzip` и повторите команду.

## Шаг 3. Проверить на сайте
Обновите страницу (Ctrl+F5, на телефоне — закрыть и открыть заново): русский язык, иконка «D»,
отправка PDF-скана, DeepSeek с pravo, строка «Скиллы / Инструменты» под ответом, кнопка ↕.

## Если что-то сломалось — откат на прежнюю версию
Возвращает старый сайт, старый сервер и базу из копии, сделанной в шаге 2.
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && V=/var/lib/docker/volumes/deeix-chat-app-data && $C stop app && rm -rf $V/_data && mkdir $V/_data && tar -xzf /opt/deeix-ru/baza-pered-0.4.6.tar.gz -C $V/_data && cp -a /opt/deeix-ru/deeix-chat-old /opt/deeix-ru/deeix-chat && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-old /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

## Уборка (через пару дней, когда всё работает)
```
clear && rm -rf /opt/deeix-ru/out-old /opt/deeix-ru/deeix-chat-old /opt/deeix-ru/baza-pered-0.4.6.tar.gz /opt/deeix-ru/deeix-ru-out.zip /opt/deeix-ru/deeix-ru-backend.zip
```
Ночная автокопия базы (`/opt/deeix-backup`, 04:00 UTC) работает как раньше.

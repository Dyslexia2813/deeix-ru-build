# Установка: просмотр «Что увидит модель» (PR #28)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/28
Сборка сайта и бэкенда: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/43614172557 (артефакты `deeix-ru-out` и `deeix-ru-backend`).
Сборка прокси: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/43614162232 (артефакт `deeix-pd-mask-image`).

Меняются три вещи: **прокси `deeix-pd-mask`** (новый образ), **сайт (`out`)** и **бэкенд (`deeix-chat`)**. База, настройки прокси (`pd-mask.env`, маршруты, секрет) и JSON моделей не меняются. В сборке сайта и бэкенда всё, что уже стоит на сервере (до PR #27 включительно), плюс эта правка.

## Шаг 1. Скачать
1. По обеим ссылкам на сборки (дождитесь зелёных галочек) внизу в разделе Artifacts скачайте три файла: `deeix-ru-out`, `deeix-ru-backend` и `deeix-pd-mask-image`.
2. В 1Panel → «Файлы» удалите старые zip с такими именами в `/opt/deeix-ru/` и загрузите новые. Если браузер добавил к имени ` (1)`, переименуйте.

## Шаг 2. Прокси
Загружает новый образ и перезапускает прокси с прежними настройками.

```
clear && cd /opt/deeix-ru && docker tag deeix-pd-mask:latest deeix-pd-mask:pered-28 && unzip -o -q deeix-pd-mask-image.zip && gunzip -c deeix-pd-mask-image.tar.gz | docker load && rm -f deeix-pd-mask-image.tar.gz && docker compose -f compose.pd.yml up -d --force-recreate && sleep 3 && docker exec deeix-pd-mask /deeix-pd-mask healthcheck && echo "Прокси: работает"
```

## Шаг 3. Сайт и бэкенд
Сохраняет текущий сайт в `out-pered-28`, бинарник в `deeix-chat-pered-28`, ставит новые и перезапускает DEEIX.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out-pered-28 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-28 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-28 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c pd-mask/preview /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `новый код:` с числом больше 0 и в конце `{"status":"ok",...}`. Если `новый код: 0`, в zip старый бинарник: скачайте его заново.

## Шаг 4. Проверить
Обновите страницу (Ctrl+F5) и откройте чат с моделью, у которой есть «Маскировка ПД».
1. Включите «Маскировку ПД». Рядом с кнопками в поле ввода появится глаз «Просмотр». Нажмите его, он подсветится (выбор запоминается в браузере).
2. Напишите: «Подготовь претензию от имени Иванова Ивана Ивановича (ИНН 500100732259, тел. +7 916 123-45-67). Директор Сидоренко подписал акт» и отправьте. Вместо отправки откроется «Что увидит модель»: сообщение с метками, таблица замен, жёлтая подсказка про «Сидоренко».
3. Нажмите «Спрятать» у «Сидоренко»: в тексте появится метка `[FIO_…]`, в таблице новая строка. Выключите «Прятать» у телефона: он вернётся в текст. Нажмите «Отправить».
4. Журнал: `clear && docker logs --tail 5 deeix-pd-mask`. Будут строки `preview texts=…` (только числа) и `masked=…` для самого запроса.
5. Без «Просмотра» (глаз выключен) отправка идёт сразу, как раньше. При выключенной «Маскировке ПД» глаза нет.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-28 /opt/deeix-ru/out && cp -a /opt/deeix-ru/deeix-chat-pered-28 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app && docker tag deeix-pd-mask:pered-28 deeix-pd-mask:latest && cd /opt/deeix-ru && docker compose -f compose.pd.yml up -d --force-recreate
```

## Уборка после проверки
```
clear && rm -rf /opt/deeix-ru/out-pered-28 /opt/deeix-ru/deeix-chat-pered-28 /opt/deeix-ru/deeix-pd-mask-image.zip && docker rmi deeix-pd-mask:pered-28
```

## Как это устроено
- Окно спрашивает прокси (через бэкенд DEEIX, `POST /api/v1/pd-mask/preview`), поэтому показывает ровно ту маскировку, которая будет при отправке. Прокси ничего не хранит.
- Выбор («не прятать телефон», «спрятать Сидоренко») сохраняется в параметрах чата и действует на следующие сообщения этого чата, пока его не поменять в окне.
- Подсказки — это догадки: слово с заглавной буквы с фамильным окончанием («-енко», «-ов», «-ин», «-ский»…) и 10/12 цифр рядом со словом «ИНН», у которых не сошлась контрольная сумма.
- Картинки и файлы без текста в окне помечены: их содержимое модель получит как есть.

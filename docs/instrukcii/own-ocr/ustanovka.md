# Установка: своё OCR (PR #19)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/19

Ставятся две вещи:
- **контейнер `deeix-ocr`**: Tesseract (русский и английский), сборка https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42193379095, артефакт `deeix-ocr-image`;
- **новый бэкенд** (бинарник `deeix-chat`), сборка https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/42196119609, артефакт `deeix-ru-backend`. В нём всё, что уже стоит на сервере (v0.4.6 со всеми патчами), плюс распознавание по требованию: OCR запускается только тогда, когда картинку или скан отправляют модели, которая не понимает картинки. Модели с картинками получают их как раньше, а при загрузке файлов сервер ничего не распознаёт. Сайт (`out`) и база не меняются.

Сколько это стоит: в простое около 5–10 МБ памяти, во время распознавания около 120 МБ (лимит контейнера 300 МБ). На диске 187 МБ под образ (скачать нужно около 75 МБ). Распознаётся одна задача за раз, с самым низким приоритетом.

## Шаг 1. Скачать и загрузить
1. По ссылке на первую сборку внизу (раздел Artifacts) скачайте `deeix-ocr-image`, придёт `deeix-ocr-image.zip`.
2. По ссылке на вторую сборку скачайте `deeix-ru-backend`, придёт `deeix-ru-backend.zip`.
3. В 1Panel откройте «Файлы» → `/opt/deeix-ru/`. Удалите старый `deeix-ru-backend.zip` и загрузите оба новых zip (`deeix-ru-backend.zip` весит около 24,3 МБ, `deeix-ocr-image.zip` около 72 МБ). Если браузер добавил к имени « (1)», уберите это.

## Шаг 2. Одна команда в терминале 1Panel
Команда загружает образ, создаёт `compose.ocr.yml`, запускает OCR, сохраняет текущий бинарник как `deeix-chat-pered-19`, ставит новый и перезапускает DEEIX.

```
clear && cd /opt/deeix-ru && unzip -o -q deeix-ocr-image.zip && gunzip -c deeix-ocr-image.tar.gz | docker load && rm -f deeix-ocr-image.tar.gz && printf 'services:\n  ocr:\n    container_name: deeix-ocr\n    image: deeix-ocr:latest\n    restart: unless-stopped\n    mem_limit: 300m\n    memswap_limit: 300m\n    cpus: 1.0\n    pids_limit: 64\n    networks: [deeix-chat]\nnetworks:\n  deeix-chat:\n    external: true\n    name: deeix-chat-network\n' > compose.ocr.yml && docker compose -f compose.ocr.yml up -d && sleep 3 && docker exec deeix-ocr deeix-ocr healthcheck && echo "OCR: работает" && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-19 && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c 'Text recognized by OCR' /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно получиться три строки: `OCR: работает`, `новый код: 1` и в конце `{"status":"ok",...}`. Если `новый код: 0`, значит, в zip лежит старый бинарник: скачайте его заново.

## Шаг 3. Включить в админке
Админка → настройки файлов → «Извлечение содержимого». Поля OCR появляются только при включённой галочке, поэтому настройка идёт в два захода:
1. «Движок извлечения» оставьте «Встроенный».
2. Временно включите **OCR изображений**. Ниже появятся поля OCR.
3. **Движок OCR**: Tesseract. **Адрес сервиса Tesseract OCR**: `http://deeix-ocr:8004/ocr`. **Тайм-аут Tesseract OCR**: `600`. Сохраните.
4. Выключите **OCR изображений** обратно (**OCR для PDF как запасной путь** тоже должна быть выключена) и сохраните ещё раз. Поля OCR скроются, но значения останутся. Это и есть режим «по требованию». Если галочки оставить включёнными, DEEIX будет распознавать каждую загрузку для всех моделей, а модели с картинками будут получать текст скана вместо самих страниц.

У моделей, которые не понимают картинки (например, DeepSeek), во входных данных (модальностях) должен стоять только «текст». Только тогда они получают распознанный текст. Если входные данные не заданы, DEEIX по-прежнему отправляет модели саму картинку.

## Шаг 4. Проверить
1. В новом чате с DeepSeek прикрепите фото или скриншот с текстом и спросите, что там написано. Ответ задержится на 10–20 секунд, пока идёт распознавание. В следующих сообщениях этого диалога задержки уже не будет: текст запоминается.
2. Прикрепите PDF-скан без текста и спросите DeepSeek о содержании. Подойдут и сканы, загруженные раньше: распознавание идёт при отправке, а не при загрузке. На длинном скане ответ может ждать несколько минут (до 20 страниц).
3. Отправьте ту же картинку модели с картинками (например, GLM): ответ придёт сразу, а в `docker logs deeix-ocr` не появится новой строки.
4. Сколько памяти занимает OCR: `clear && docker stats --no-stream deeix-ocr`. Журнал: `clear && docker logs --tail 20 deeix-ocr`.

## Откат
Выключите обе галочки OCR в админке, затем:
```
clear && docker rm -f deeix-ocr && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat-pered-19 /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```

## Уборка после проверки
```
clear && rm -f /opt/deeix-ru/deeix-chat-pered-19 /opt/deeix-ru/deeix-ocr-image.zip
```

## Что умеет и чего нет
- Печатный текст распознаётся хорошо, рукописный и сложная вёрстка плохо. Таблицы приходят строками, без сетки.
- PDF: не больше 20 страниц за файл, остальные страницы пропускаются.
- Модели с картинками (GLM и другие) по-прежнему получают саму картинку или страницы скана, OCR для них не запускается.
- Распознанный текст хранится в памяти DEEIX (до 256 файлов). После перезапуска DEEIX файл при следующей отправке распознаётся ещё раз.

# Модель создаёт файлы Word и Excel (PR #32)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/32
Сборка бэкенда: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/44782710905
Сборка образа deeix-office: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/44782726685

Ставятся две вещи: новый небольшой контейнер `deeix-office` (собирает файлы, лимит памяти 64 МБ) и новый бинарник DEEIX (добавлена ссылка для скачивания). Сайт (`out`) и база не меняются.

## Шаг 1. Скачать
1. По первой ссылке (бэкенд) в разделе Artifacts скачайте `deeix-ru-backend`. По второй ссылке (образ) скачайте `deeix-office-image`.
2. В 1Panel → «Файлы» загрузите в `/opt/deeix-ru/` оба файла: `deeix-ru-backend.zip` и `deeix-office-image.zip`. Старый `deeix-ru-backend.zip` сначала удалите. Если браузер добавил к имени « (1)», уберите это из имени.

## Шаг 2. Одна команда в терминале 1Panel
Команда загружает образ, создаёт `compose.office.yml`, запускает контейнер, сохраняет текущий бинарник как `deeix-chat-pered-office`, ставит новый и перезапускает DEEIX.

```
clear && cd /opt/deeix-ru && unzip -o -q deeix-office-image.zip && gunzip -c deeix-office-image.tar.gz | docker load && rm -f deeix-office-image.tar.gz && printf 'name: deeix-office\nservices:\n  office:\n    container_name: deeix-office\n    image: deeix-office:latest\n    restart: unless-stopped\n    environment:\n      OFFICE_TTL_DAYS: "7"\n      OFFICE_MAX_MB: "200"\n    volumes:\n      - deeix-office-data:/data\n    mem_limit: 64m\n    memswap_limit: 64m\n    cpus: 0.5\n    pids_limit: 64\n    read_only: true\n    networks: [deeix-chat]\nvolumes:\n  deeix-office-data:\n    name: deeix-office-data\nnetworks:\n  deeix-chat:\n    external: true\n    name: deeix-chat-network\n' > compose.office.yml && docker compose -f compose.office.yml up -d --force-recreate && sleep 3 && docker exec deeix-office /deeix-office healthcheck && echo "deeix-office: работает" && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat /opt/deeix-ru/deeix-chat-pered-office && unzip -o -q /opt/deeix-ru/deeix-ru-backend.zip -d /opt/deeix-ru && chmod +x /opt/deeix-ru/deeix-chat && echo "новый код: $(grep -c 'office files unavailable' /opt/deeix-ru/deeix-chat)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть `deeix-office: работает`, затем `новый код: 1` и в конце `{"status":"ok",...}`.

## Шаг 3. Подключить в DEEIX
Админка → Инструменты → MCP → Добавить сервер:
- Имя: `office`
- URL: `http://deeix-office:8095/mcp`
- Статус: включён

Нажмите **Sync**: появятся 2 инструмента, `make_docx` и `make_xlsx`. В чате в списке MCP-инструментов будет строка «office 0/2».

## Шаг 4. Проверить
Новый чат, любая модель, отметьте office и напишите:
- «Сделай в Word короткое письмо-претензию о просрочке оплаты по договору поставки, с таблицей расчёта неустойки».
- «Сделай в Excel реестр из трёх договоров: номер, контрагент, ИНН, сумма, дата; внизу итог формулой».

В ответе должна появиться ссылка «📎 ….docx» или «📎 ….xlsx». Нажмите её, и файл скачается. Ссылка работает 7 дней.

## Если что-то не так
- Модель не вызывает инструмент: попросите прямо «сделай файл Word». Проверьте, что office отмечен в чате.
- Ссылка показывает «office files unavailable»: контейнер не запущен. Проверить: `docker ps --filter name=deeix-office` и `docker logs --tail 20 deeix-office`.
- После Sync нет инструментов, или ошибка про адрес: проверьте, что контейнер запущен (`docker ps`). Если DEEIX не принимает имя `deeix-office` как адрес, напишите мне.

## Откат
```
clear && docker compose -f /opt/deeix-ru/compose.office.yml down && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && $C stop app && cp -a /opt/deeix-ru/deeix-chat-pered-office /opt/deeix-ru/deeix-chat && $C up -d --force-recreate app
```
И удалите сервер `office` в Админка → Инструменты → MCP.

## Уборка (когда всё проверено)
```
clear && rm -f /opt/deeix-ru/deeix-chat-pered-office /opt/deeix-ru/deeix-office-image.zip /opt/deeix-ru/deeix-ru-backend.zip && echo готово
```

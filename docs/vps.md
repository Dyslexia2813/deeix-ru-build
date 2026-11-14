# VPS: как устроен сервер

Состояние на 2026-11-14. Если что-то меняется на сервере — обнови этот файл.

## Машина

- Ubuntu 24.04 amd64, ядро 6.8.0-149-generic (обновлено 2026-11-13), ~709 МБ ОЗУ, swap 2 ГБ (`/swapfile` в fstab, `vm.swappiness=10`). Время сервера — UTC.
- Панель 1Panel (с 2FA). Kimi работает через её веб-терминал и раздел «Файлы». SSH снаружи только по ключу, пароль разрешён лишь с localhost (`/etc/ssh/sshd_config.d/10-no-password.conf`; так подключается веб-терминал 1Panel).
- ufw включён. Docker публикует порты в обход ufw — учитывай это.
- Реверс-прокси / TLS: контейнер `openresty` (1Panel). Cloudflare не используется.
- Сайт: `tap.teamdream.online`.
- Мало памяти: на сервере ничего не собирать (Next.js требует 2–3 ГБ). Всё собирается в GitHub Actions.

## DEEIX Chat

- Compose-проект: `/opt/deeix-chat/DEEIX-Chat/deploy`, файл `docker-compose.sqlite.yml` + наш `/opt/deeix-ru/compose.override.yml`. Команда, которую используют все инструкции:

  ```
  cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml"
  ```
- Сервис `app`, контейнер `deeix-chat-app`, образ `ghcr.io/deeix-ai/deeix-chat:latest`, порт `127.0.0.1:8080` (проверка: `curl -s http://127.0.0.1:8080/healthz` → `{"status":"ok","version":"0.4.6"...}`).
- Сеть Docker `deeix-chat-network` (172.19.0.0/16, шлюз 172.19.0.1, контейнер app = 172.19.0.2). `docker compose down` может пересоздать сеть с другим адресом — тогда сломаются MCP-мосты на 172.19.0.1. `up -d --force-recreate` сеть не трогает.
- База SQLite и загруженные файлы: том `deeix-chat-app-data` (`/var/lib/docker/volumes/deeix-chat-app-data/_data`).
- Override подменяет в образе:
  - `/opt/deeix-ru/out` → `/app/frontend/out` (наш сайт);
  - `/opt/deeix-ru/deeix-chat` → `/app/deeix-chat` (наш бинарник);
  - `restart: unless-stopped`, `mem_limit: 320m`, `memswap_limit: 1g`.
- Поэтому `docker compose pull` без нашей новой сборки по сути ничего не меняет.
- Замена бинарника: работающий файл перезаписать нельзя (`Text file busy`) — сначала `$C stop app`, потом копировать, `chmod +x`, `$C up -d --force-recreate app`.
- Версия на сервере: **v0.4.6** + все PR deeix-ru-build до #28 включительно (main = то, что стоит на сервере).

## Другие сервисы

| Сервис | Где | Адрес для DEEIX | Заметка |
|---|---|---|---|
| `pravo-bridge` (systemd) | `/opt/pravo-mcp` (venv, `bridge.py`, fastmcp 3.4.8 — **не обновлять до 4**) | `http://172.19.0.1:8765/mcp` | MCP pravo.gov.ru. В `server.py` внутри venv локальные патчи (типы документов, ИПС, редакции) — пропадут при переустановке, см. `docs/arhiv/STATE-2026-10-15.md` |
| `deepseek-search` (systemd) | `/opt/deepseek-search`, ключ `/etc/deepseek-search.env` | `http://172.19.0.1:8767/mcp` | Инструменты `ds_search`, `ds_fetch`. В DEEIX называется `deepseek-web-search`. [Инструкция](instrukcii/deepseek-search/ustanovka.md) |
| `deeix-ocr` (Docker) | `/opt/deeix-ru/compose.ocr.yml`, лимит 300m | `http://deeix-ocr:8004/ocr` | Tesseract rus+eng, только для текстовых моделей. [Инструкция](instrukcii/own-ocr/ustanovka.md) |
| `deeix-pd-mask` (Docker) | `/opt/deeix-ru/compose.pd.yml`, настройки `/opt/deeix-ru/pd-mask.env` | `http://deeix-pd-mask:8090/<маршрут>` | Маскировка ПД, маршруты `ds`, `kdk`, `zai`. [Инструкция](instrukcii/pd-mask/ustanovka.md) |
| `zai-search` | — | — | Больше не нужен (заменён deepseek-search) |
| RustDesk (`1Panel-rustdesk-Wnks`) | Docker | — | Выключен с 2026-11-01. Вернуть: `docker update --restart=always 1Panel-rustdesk-Wnks && docker start 1Panel-rustdesk-Wnks` |

MCP-мосты слушают только 172.19.0.1 (шлюз Docker-сети), ufw пускает к ним только 172.19.0.0/16. Переносить на 127.0.0.1 нельзя — контейнер DEEIX до них не дотянется.

## Резервные копии

- Ночная копия базы: `/opt/deeix-backup/backup.sh`, `/etc/cron.d/deeix-backup`, 04:00 UTC, хранит 7 штук, журнал `/opt/deeix-backup/backup.log`.
- Ручная копия и восстановление: [instrukcii/db-backup/kopiya-bd.md](instrukcii/db-backup/kopiya-bd.md).
- Копии лежат на том же VPS — Kimi стоит периодически скачивать их к себе.

## Безопасность (история)

- Таблица nftables `ip cve_2026_98432` (юнит `cve-2026-98432.service`) оставлена, не мешает. MPTCP выключен (`/etc/sysctl.d/99-no-mptcp.conf`). Подробности: [zametki/vps-cve-2026-98432.md](zametki/vps-cve-2026-98432.md).
- Другие разборы CVE: `zametki/vps-cve-*.md`.

## Настройки DEEIX, которые полезно знать

- Лимиты инструментов (Админка → Инструменты): по умолчанию 8 MCP-вызовов и 5 запросов к LLM на сообщение; таймаут MCP поднят до 60 с.
- Модель видна в чате, только если у неё есть цена в «Цены моделей» (цена привязана к имени модели) — [zametki/deeix-model-catalog.md](zametki/deeix-model-catalog.md).
- Повторная загрузка того же файла берётся из уже загруженных (дедуп по SHA-256), кроме файлов со статусом failed. Чтобы файл перечитался новым кодом, его надо удалить в «Файлах» и загрузить заново.

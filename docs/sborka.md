# Сборка, патчи и обновление DEEIX

## Что лежит в репозитории

```
deeix-ru-addon.tar.gz      исходники аддона (архивом, см. ниже)
.github/workflows/
  build.yml                сайт + бэкенд DEEIX с аддоном (вручную, вход tag)
  ocr-image.yml            образ deeix-ocr (на PR с изменениями в ocr-service/)
  pd-mask-image.yml        образ deeix-pd-mask (на PR с изменениями в pd-mask/)
ocr-service/               своё OCR: Go + Tesseract rus+eng
pd-mask/                   прокси маскировки персональных данных (Go), см. pd-mask/README.md
docs/                      документация проекта
```

## Аддон `deeix-ru-addon.tar.gz`

Внутри папка `deeix-ru-addon/`:

- `install.sh [--check | --uninstall] [--no-<патч> ...] <клон deeix-chat>` — накладывает патчи через `git apply` (идемпотентно; если патч не ложится, значит апстрим поменял файл) и копирует `files/` в исходники.
- `patches/*.patch` — правки существующих файлов апстрима. Порядок задан массивом `PATCHES` в `install.sh`. По отдельности `--check` может ругаться на патчи, которые зависят от предыдущих (`turn-resources`, `focus-toggle` и т.п.) — это нормально, важен проход полного `install.sh`.
- `files/` — новые файлы (русские переводы `apps/web/i18n/messages/ru-RU/*.json`, загрузчик `ru-messages.ts`, `ru-server-text.ts` — перевод китайских текстов сервера при показе, экспорт PDF/DOCX и т.д.).
- `tools/` — проверка переводов: `missing-keys.mjs` (непереведённые ключи), `check-ru.mjs`, `merge-parts.mjs`, `GLOSSARY.ru.md` (глоссарий).
- `README.md` — описание каждого патча, раздел «Конфликты» (что делать, если патч не ложится).

### Как вносить правку

```
tar -xzf deeix-ru-addon.tar.gz                       # получаем deeix-ru-addon/
git clone --depth 1 --branch v0.4.6 https://github.com/DEEIX-AI/deeix-chat /tmp/src
bash deeix-ru-addon/install.sh /tmp/src              # всё наше поверх чистого тега
# правим код в /tmp/src, затем делаем патч из разницы только по своим файлам:
#   cd /tmp/src && git diff -- <файлы> > ../deeix-ru-addon/patches/<имя>.patch
#   (если файл уже правит другой патч — патч делается поверх состояния после предыдущих патчей)
# добавляем ключ --no-<имя> и строку в PATCHES в install.sh, раздел в README.md
tar -czf deeix-ru-addon.tar.gz deeix-ru-addon && rm -rf deeix-ru-addon
```

Правила:

- Новая правка — отдельный патч с ключом `--no-<имя>`, чтобы её можно было выключить.
- Русские строки — в `files/.../ru-RU/*.json`, английские и китайские — внутри патча (в en/zh файлах апстрима). Подписи можно брать через `useLocale` в коде, не трогая json (так сделано в focus-toggle, sidebar-width).
- Без новых npm-зависимостей (решение Kimi: чтобы аддон не ломался при обновлениях).
- Китайские константы сервера в `message-process-trace.ts` нужны для распознавания этапов — переводить только при показе.
- В SQLite `LOWER()` не понижает кириллицу — учитывай в запросах и триггерах скиллов.

### Проверка перед PR

- `install.sh` на чистый клон тега проходит; `install.sh --uninstall` снимает.
- Бэкенд: `cd backend && go build ./... && go test ./...` (v0.4.6 требует Go 1.26: `GOTOOLCHAIN=auto`). Апстрим-тест `TestApplicationInfraAdapterRatchet` запрещает зависимость application → infra.
- Фронтенд: `pnpm install && pnpm --filter @deeix/web check` (tsc, biome, архитектурные правила: без необъяснённых `as`, `shared` не зависит от `entities`/`features`, `entities` не зависит от `features`, хуки чата — `features/chat/hooks/use-chat-*.ts`).
- Для интерфейса — посмотреть в браузере (Playwright, ПК и телефон 390 px).

### Локальный запуск для проверки в браузере

- Бэкенд: `go build -o /tmp/deeix-server ./cmd/server` в `backend`; конфиг — из `deploy/config.sqlite.example.yaml`, пути поменять, `frontend_dist_dir` на `apps/web/out` (сайт: `next build`).
- Пароль superadmin — в логе первого запуска («bootstrap superadmin created»). Онбординг убрать: `PATCH /api/v1/me/username`, `POST /api/v1/me/onboarding/complete {newPassword}` (или пройти в UI / спрятать модалку CSS в Playwright).
- Токен: `POST /api/v1/auth/login {username,password}` → `data.accessToken` (Bearer).
- Тестовые данные: чаты `POST /api/v1/conversations {title}`, проекты `POST /api/v1/conversation-projects {name}`, промпты `POST /api/v1/prompt-presets/mine`, готовые ответы с журналом можно вставить прямо в SQLite (`chat_messages`, `chat_runs`, `chat_run_events`). Чат открывается `/chat?conversation_id=<public_id>`.
- Биллинг для проверки цен: `PATCH /admin/billing/config {"mode":"usage"}`; язык профиля: `PATCH /me {"locale":"ru-RU"}`.
- В Playwright для полос прокрутки: `ignoreDefaultArgs: ['--hide-scrollbars']`.

## Сборка в GitHub Actions

- `build.yml`: вручную (Actions → Run workflow, или `gh workflow run build.yml --ref <ветка> -f tag=v0.4.6`). Клонирует апстрим по тегу, ставит аддон, собирает официальным Dockerfile (`frontend-builder`, `backend-builder`). Артефакты: `deeix-ru-out` (папка сайта) и `deeix-ru-backend` (бинарник linux/amd64).
- Тег должен совпадать с версией на сервере (`/healthz`).
- Архив надо **закоммитить** — иначе workflow собирает старый.
- Если в двух PR правился архив, при слиянии второго будет конфликт: обнови его ветку слиянием main, пересобери архив и сборку.
- Чтобы убедиться, что на сервер попал новый бинарник, в инструкции есть `grep -c <уникальная строка> /opt/deeix-ru/deeix-chat` («новый код: N», должно быть > 0). Для сайта — искать строку в `out/_next`.

## Обновление DEEIX на новую версию апстрима

1. Новые теги: https://github.com/DEEIX-AI/deeix-chat/releases (уведомление в админке только ссылается на релиз).
2. Клон нового тега, `install.sh --check`, затем полный `install.sh`. Если патч не ложится — перенести: клон с двумя тегами, наши патчи коммитами на старом теге, `cherry-pick` на новый, затем `go build/test` и `pnpm check` (чистый cherry-pick не гарантирует сборку).
3. Новые строки апстрима: `node deeix-ru-addon/tools/missing-keys.mjs <клон> <папка частей> <куда писать>` (подробно — раздел «Инструменты» в `deeix-ru-addon/README.md`) → перевести. `merge-parts.mjs` перезаписывает ru-файл только из частей — старый ru-файл класть как часть `.0`.
4. Поменять `default: "v0.4.x"` в `build.yml`, PR, сборка.
5. Инструкция для Kimi: по образцу [instrukcii/deeix-update/obnovlenie-v0.4.6.md](instrukcii/deeix-update/obnovlenie-v0.4.6.md) — **сначала копия базы**, `$C pull app`, копии старого `out` и бинарника, установка, откат с восстановлением базы.

История переходов: [zametki/deeix-upgrade-v045.md](zametki/deeix-upgrade-v045.md), [zametki/deeix-upgrade-v046.md](zametki/deeix-upgrade-v046.md).

# deeix-data-table-fit

_Таблицы-реестры (компонент data-table) — перенос текста и кнопка «Открыть во всю ширину» вместо боковой панели — patches/data-table-fit.patch, PR #24._

- 2026-11-08 Kimi: в «реестре» (интерактивная таблица: «Строк: N», поиск, скачать, сортировка) текст в одну строку, таблица бесконечно широкая; кнопка «Открыть в боковой панели» бесполезна — хочет «во всю ширину».
- Причина: [deeix-table-fit](deeix-table-fit.md) правит только markdown-таблицы; data-table (apps/web/shared/components/markdown/ui-blocks/data-table.tsx) использует components/ui/table.tsx с whitespace-nowrap и min-w-max.
- patches/data-table-fit.patch (ключ --no-data-table-fit, ставится последним): столбцы с текстом ≥24 симв. переносятся (min 9rem, ≥90 симв. 12–14rem), короткие nowrap; sticky первого столбца только если он короткий; ui-block-host.tsx — UIBlockOpenable открывает Dialog на весь экран (UIBlockFullscreenProvider в ui-block-actions.tsx, подписи через useLocale). Виджеты deeix-widget остались с боковой панелью.
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/24 (Kimi проверил на VPS, слит 2026-11-08, squash 3ffa576, стоит на сервере), сборка run 42496055924, инструкция и скриншоты ../instrukcii/data-table-fit/ (копия out-pered-reestrom).
- Локальная проверка: сервер /tmp/deeix-server из backend/cmd/server, config.yaml в cwd; пароль admin в логе «bootstrap superadmin created»; PATCH /me/username, POST /me/onboarding/complete {newPassword}; сообщение с ```deeix-ui вставлять прямо в chat_messages (status completed); чат открывается /chat?conversation_id=...

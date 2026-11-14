# deeix-upgrade-v045

_Переход сборки DEEIX-RU с v0.4.4-beta.4 на v0.4.5 (2026-10-20): что изменилось в патчах, PR #5, как обновлять апстрим дальше._

- 2026-10-20 вышел DEEIX v0.4.5. Уведомление в админке только ссылается на релиз, само ничего не обновляет. Образ в compose = ghcr.io/deeix-ai/deeix-chat:latest, но override монтирует наш бинарник и out, поэтому `pull` без новой сборки ничего не меняет по сути.
- Все 4 патча нужны и на v0.4.5 (апстрим не исправил DeepSeek/Anthropic; добавил свою отправку «PDF без текста» целиком только моделям, читающим PDF — Claude/Gemini/OpenAI).
- Перенос: пакет pdfpageimages перенесён в backend/internal/pkg (апстрим-тест TestApplicationInfraAdapterRatchet запрещает application→infra); тесты anthropic_dsml под новую сигнатуру buildAnthropicContent(msg, nil); в resolve-error-message.ts добавлен ru-RU в ERROR_REFERENCE_TEMPLATES. Переведены новые строки (33 файла ru-RU, новые admin-user-keys, desktop-downloads).
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/5 (сборка run 39498805066) слит 2026-10-22. Инструкция для VPS: ../instrukcii/deeix-update/obnovlenie-v0.4.5.md (копии старой версии в /opt/deeix-ru/*-0.4.4, удалить после проверки). Установку на VPS Kimi явно не подтверждал.
- Способ переноса на новую версию: репо из двух тегов, патчи коммитами на старом теге, cherry-pick на новый, затем `go build/test ./...` (GOTOOLCHAIN=auto) и `pnpm --filter @deeix/web check` — чистый cherry-pick не гарантирует сборку. Инструмент tools/merge-parts.mjs перезаписывает ru-файл только из частей: класть старый ru-файл как часть .0.
- См. [vps-memory-scan-pdf](vps-memory-scan-pdf.md), [deeix-deepseek-anthropic](deeix-deepseek-anthropic.md), [deeix-app-icon](deeix-app-icon.md).

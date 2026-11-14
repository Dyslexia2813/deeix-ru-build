# deeix-server-text-ru

_Китайские тексты сервера DEEIX переводятся при показе (ru-server-text.ts) и ошибка 500 статистики на SQLite (usage-stats.patch) — PR #15._

- 2026-10-31 Kimi: «请求未完成» под «Обработка завершена» и «Не удалось загрузить статистику использования — внутренняя ошибка сервера».
- Сервер пишет тексты трассы, ошибок запросов, обработки файлов/OCR, статусы сервисов по-китайски и хранит их в БД. Решение: files/apps/web/i18n/ru-server-text.ts (localizeServerText, по document.documentElement.lang = ru), подключён в message-process-trace.ts, message-agent-tool-step.tsx, file-processing.ts, content-meta.tsx, settings-runtime-panel.tsx, admin-tools.tsx (всё в ru-locale.patch). Письма с кодом подтверждения переведены в registration.go. Китайские константы в message-process-trace.ts нужны для распознавания этапов — переводить только при показе.
- Переведены admin-files, admin-upstreams, остаток admin-models, desktop-network: missing-keys пусто.
- Статистика: на SQLite ROUND(AVG(latency_ms)) — float; при среднем ≥ 1 000 000 мс Go читает "1e+06" в int64 → 500. Фикс CAST(... AS BIGINT) в 3 запросах billing/repository.go, patches/usage-stats.patch (--no-usage-stats), тест TestUsageStatisticsReadsLargeAverageLatencyOnSQLite. Код ошибки из окна можно искать в админке → журнал ошибок.
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/15 (слит 2026-11-03, 484e726), сборка run 41344704467 (после слияния с #16), инструкция ../instrukcii/server-text-ru/ustanovka.md (сайт + бэкенд). Kimi поставил поверх #16 (благодарил 2026-11-03).
- В этой песочнице Go работает через GOTOOLCHAIN=auto (скачивает 1.26.8). См. [deeix-tools-not-called](deeix-tools-not-called.md), [deeix-upgrade-v045](deeix-upgrade-v045.md).

# deeix-pro-hang

_deepseek-v4-pro в DEEIX «зависал» 2026-11-03 (молчание, ↓0) — затор на стороне DeepSeek, прошёл сам; как быстро отличить, кто виноват._

- 2026-11-03 ~09:15–10:45 UTC deepseek-v4-pro (источник «DeepSeek /anthropic») молчал 20–180 с на любой запрос, Kimi останавливал вручную; flash работал. К 11:10 UTC всё заработало само.
- Доказательство, что не DEEIX: curl с VPS напрямую к api.deepseek.com (и /anthropic, и OpenAI, stream:true) → 200, затем только ": keep-alive" 180 с (DeepSeek держит запрос в очереди). Тот же ключ, flash в это время отвечал.
- В БД такие запуски: chat_runs.status=canceled, error пустой (остановлено пользователем). «Запрос не выполнен» в трассе = есть ошибка/отмена (traceErrorSummary, service_trace.go).
- Готовые проверки: ../instrukcii/deepseek-pro-hang/ — diagnostika.md (БД, только чтение), curl-test.md (stream через оба входа), models-test.md (GET /models + «привет» каждой модели). Имена у ключа Kimi: deepseek-flash (V4.1-Flash), deepseek-v4-pro.
- Попутные изменения в JSON возможностей pro (оставлены): defaultOptions {max_tokens 393216, thinking enabled} (без него DEEIX шлёт thinking adaptive) и уровень "none" в reasoning.levels (кнопка «Выкл.»).
- См. [deeix-deepseek-search](deeix-deepseek-search.md), [deeix-masking-timeout](deeix-masking-timeout.md).

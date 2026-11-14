# Проверка deepseek-v4-pro напрямую, без DEEIX

Вставьте свой ключ DeepSeek вместо `sk-ВАШ_КЛЮЧ` и выполните целиком в терминале 1Panel.
Команда задаёт «привет» дважды: через вход /anthropic (как в DEEIX) и через обычный вход OpenAI (как обычно в opencode),
и печатает, через сколько секунд пришёл первый байт и весь ответ.

```
clear && K="sk-ВАШ_КЛЮЧ" && \
echo "== /anthropic" && curl -sS -o /tmp/ds-a.txt -w "первый байт: %{time_starttransfer} c, всего: %{time_total} c, код %{http_code}\n" --max-time 180 https://api.deepseek.com/anthropic/v1/messages -H "x-api-key: $K" -H "anthropic-version: 2023-06-01" -H "content-type: application/json" -d '{"model":"deepseek-v4-pro","max_tokens":200,"stream":true,"thinking":{"type":"disabled"},"messages":[{"role":"user","content":"привет"}]}' ; head -c 400 /tmp/ds-a.txt; echo; \
echo "== OpenAI" && curl -sS -o /tmp/ds-o.txt -w "первый байт: %{time_starttransfer} c, всего: %{time_total} c, код %{http_code}\n" --max-time 180 https://api.deepseek.com/chat/completions -H "Authorization: Bearer $K" -H "content-type: application/json" -d '{"model":"deepseek-v4-pro","max_tokens":200,"stream":true,"thinking":{"type":"disabled"},"messages":[{"role":"user","content":"привет"}]}' ; head -c 400 /tmp/ds-o.txt; echo
```

# Какие имена моделей отдаёт DeepSeek и какое из «pro» отвечает

Вставьте ключ вместо `sk-ВАШ_КЛЮЧ`, выполните целиком. Каждое имя ждём не дольше 40 с.

```
clear && K="sk-ВАШ_КЛЮЧ" && echo "== список моделей" && curl -sS https://api.deepseek.com/models -H "Authorization: Bearer $K"; echo; \
for M in $(curl -sS https://api.deepseek.com/models -H "Authorization: Bearer $K" | grep -o '"id":"[^"]*"' | cut -d'"' -f4) deepseek-pro; do \
printf "== %s: " "$M"; curl -sS -o /tmp/ds-m.txt -w "%{http_code}, %{time_total} c\n" --max-time 40 https://api.deepseek.com/chat/completions -H "Authorization: Bearer $K" -H "content-type: application/json" -d "{\"model\":\"$M\",\"max_tokens\":30,\"thinking\":{\"type\":\"disabled\"},\"messages\":[{\"role\":\"user\",\"content\":\"привет\"}]}"; head -c 200 /tmp/ds-m.txt; echo; done
```

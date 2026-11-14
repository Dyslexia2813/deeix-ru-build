# Диагностика зависания deepseek-v4-pro

Вставить целиком в терминал 1Panel (только чтение базы, ничего не меняет):

```
clear && python3 - <<'EOF'
import sqlite3, json
db = sqlite3.connect("file:/var/lib/docker/volumes/deeix-chat-app-data/_data/deeix.db?mode=ro", uri=True)
db.row_factory = sqlite3.Row
q = lambda s, *a: db.execute(s, a).fetchall()
print("== последние запуски deepseek-v4-pro")
runs = q("select id, run_id, started_at, status, error_code, error_message, upstream_name, upstream_model_name, input_tokens, output_tokens, reasoning_tokens, reasoning_effort, first_token_latency_ms, total_latency_ms from chat_runs where platform_model_name like '%pro%' order by id desc limit 4")
for r in runs: print(dict(r))
print("== настройки модели")
for m in q("select id, name, capabilities_json from llm_platform_models where name like '%deepseek%'"):
    print(m["id"], m["name"], m["capabilities_json"][:1500])
print("== маршруты и источники")
for r in q("select m.name, r.protocol, r.status, um.upstream_model_name, u.name un, u.base_url, u.read_timeout_ms, u.stream_idle_timeout_ms from llm_model_routes r join llm_platform_models m on m.id=r.platform_model_id join llm_upstream_models um on um.id=r.upstream_model_id join llm_upstreams u on u.id=um.upstream_id where m.name like '%deepseek%'"):
    print(dict(r))
if runs:
    print("== события последнего запуска (ошибки)")
    for e in q("select event_type, stage, status, title, summary, substr(error_json,1,1500) err, substr(payload_json,1,600) p from chat_run_events where run_id=? and (error_json<>'' or status in ('failed','error') or event_type like '%error%') order by seq", runs[0]["run_id"]):
        print(dict(e))
EOF
```

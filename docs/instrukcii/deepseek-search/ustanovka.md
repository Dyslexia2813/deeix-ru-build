# Веб-поиск DeepSeek в DEEIX: отдельный MCP-сервер «deepseek»

Отдельный сервер `deepseek` (не внутри pravo) с двумя инструментами:
- `ds_search` — поиск в интернете (DeepSeek Search API), у каждого результата готовая выжимка;
- `ds_fetch` — прочитать страницу по ссылке целиком (DeepSeek Fetch API), длинные частями.

Работает с любой моделью DEEIX (GLM, DeepSeek и др.) и вместе с pravo.
Оплата — с баланса DeepSeek: по сторонним данным ~$0,75 за 1000 успешных запросов (неудачные бесплатны),
и, по слухам, бесплатно до 1 ноября 2026. Сверить в кабинете DeepSeek (platform.deepseek.com → Usage).

- Работает на Python от pravo (`/opt/pravo-mcp/venv`), доустанавливать ничего не нужно.
- Слушает `172.19.0.1:8767/mcp`, только для контейнера DEEIX, наружу не открыт.
- Ключ DeepSeek хранится в `/etc/deepseek-search.env` (доступ только root). Подойдёт тот же ключ, что у моделей DeepSeek.

## Шаг 1. Установка на VPS

Скопировать **весь** блок ниже (до последней строки `SVCEOF` включительно) и вставить в веб-терминал 1Panel.
Он спросит ключ DeepSeek (не отображается при вводе). В конце должно быть `active` и `ответ моста: 406`.

```bash
clear && read -s -p "Ключ DeepSeek: " K && echo && printf 'DEEPSEEK_API_KEY=%s\n' "$K" > /etc/deepseek-search.env && chmod 600 /etc/deepseek-search.env && mkdir -p /opt/deepseek-search && cat > /opt/deepseek-search/deepseek_search.py <<'PYEOF' && cat > /etc/systemd/system/deepseek-search.service <<'SVCEOF' && systemctl daemon-reload && systemctl enable --now deepseek-search && ufw allow from 172.19.0.0/16 to 172.19.0.1 port 8767 proto tcp && sleep 4 && systemctl is-active deepseek-search && curl -s -o /dev/null -w 'ответ моста: %{http_code} (406 = норма)\n' http://172.19.0.1:8767/mcp
"""MCP-сервер «deepseek» для DEEIX: веб-поиск и чтение страниц через DeepSeek Search/Fetch API.

Даёт любой модели в DEEIX два инструмента: ds_search и ds_fetch.
Запуск: DEEPSEEK_API_KEY=... python deepseek_search.py  (слушает DS_HOST:DS_PORT, путь /mcp)
"""

import os

import httpx
from fastmcp import FastMCP

BASE_URL = os.environ.get("DEEPSEEK_BASE_URL", "https://api.deepseek.com").rstrip("/")
API_KEY = os.environ.get("DEEPSEEK_API_KEY", "")
HOST = os.environ.get("DS_HOST", "172.19.0.1")
PORT = int(os.environ.get("DS_PORT", "8767"))
MAX_QUERY_CHARS = 128

mcp = FastMCP("deepseek")


async def _post(path: str, body: dict, timeout: float):
    if not API_KEY:
        return None, {"error": "на сервере не задан DEEPSEEK_API_KEY"}
    try:
        async with httpx.AsyncClient(timeout=timeout) as client:
            resp = await client.post(
                BASE_URL + path, json=body, headers={"Authorization": f"Bearer {API_KEY}"}
            )
    except httpx.HTTPError as exc:
        return None, {"error": f"DeepSeek недоступен: {exc.__class__.__name__}"}
    try:
        data = resp.json()
    except ValueError:
        data = None
    if resp.status_code != 200:
        detail = data.get("error") if isinstance(data, dict) else None
        return None, {
            "error": f"DeepSeek ответил {resp.status_code}",
            "details": detail or resp.text[:500],
        }
    return data, None


@mcp.tool
async def ds_search(query: str, count: int = 5) -> dict:
    """Поиск в интернете (DeepSeek Search). Используй для любых свежих или внешних сведений:
    новости, курсы валют, цены, события, сведения о компаниях и людях — всё, чего нет в других
    инструментах. query — короткий поисковый запрос (до 128 символов); count — сколько
    результатов (1–10). У каждого результата есть summary — обычно его хватает для ответа;
    если нужен полный текст страницы, вызови ds_fetch с её url. В ответе указывай источники ссылками."""
    query = " ".join((query or "").split())
    if not query:
        return {"error": "пустой запрос"}
    if len(query) > MAX_QUERY_CHARS:
        query = query[:MAX_QUERY_CHARS].strip()
    count = max(1, min(int(count or 5), 10))
    data, err = await _post(
        "/v1/tools/search", {"query": query, "count": count, "need_content": False}, 30
    )
    if err:
        return err
    entries = data.get("results") if isinstance(data, dict) else None
    results = []
    for item in entries or []:
        if not isinstance(item, dict) or not str(item.get("url", "")).startswith("http"):
            continue
        results.append(
            {
                "title": item.get("title") or item.get("url"),
                "url": item.get("url"),
                "site": item.get("site_name") or "",
                "date": (item.get("date") or "")[:10],
                "summary": item.get("summary") or item.get("snippet") or "",
            }
        )
    return {"query": query, "results": results}


@mcp.tool
async def ds_fetch(url: str, offset: int = 0, max_chars: int = 12000) -> dict:
    """Прочитать страницу по ссылке (DeepSeek Fetch): возвращает заголовок и текст в Markdown.
    Используй после ds_search, когда summary не хватает. Длинный текст отдаётся частями:
    offset — с какого символа читать, max_chars — сколько (до 30000); в ответе next_offset,
    если текст не закончился. Некоторые сайты DeepSeek не читает — тогда отвечай по summary."""
    url = (url or "").strip()
    if not url.startswith(("http://", "https://")):
        return {"error": "нужна ссылка http(s)"}
    data, err = await _post("/v1/tools/fetch", {"url": url[:8192]}, 50)
    if err:
        return err
    content = data.get("content") if isinstance(data, dict) else None
    if not isinstance(content, str) or not content.strip():
        return {"error": "страница пустая или не прочиталась"}
    offset = max(0, int(offset or 0))
    max_chars = max(1000, min(int(max_chars or 12000), 30000))
    part = content[offset : offset + max_chars]
    result = {
        "url": url,
        "title": data.get("title") or "",
        "total_chars": len(content),
        "content": part,
    }
    if offset + max_chars < len(content):
        result["next_offset"] = offset + max_chars
    return result


if __name__ == "__main__":
    mcp.run(transport="http", host=HOST, port=PORT, path="/mcp")
PYEOF
[Unit]
Description=DeepSeek search MCP for DEEIX
After=network-online.target docker.service
Wants=network-online.target

[Service]
DynamicUser=yes
StateDirectory=deepseek-search
Environment=HOME=/var/lib/deepseek-search
EnvironmentFile=/etc/deepseek-search.env
ExecStart=/opt/pravo-mcp/venv/bin/python /opt/deepseek-search/deepseek_search.py
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVCEOF
```

## Шаг 2. Подключить в DEEIX

Админка → Инструменты → MCP → Добавить сервер:
- Имя: `deepseek`
- URL: `http://172.19.0.1:8767/mcp`
- Статус: включён

Нажать **Sync**: появятся 2 инструмента. В чате в списке MCP-инструментов будет отдельная строка «deepseek 0/2».

## Шаг 3. Проверка

Новый чат → любая модель → отметить deepseek (и pravo, если нужно) → «Какой курс доллара ЦБ РФ на сегодня? Укажи источник».
Должен быть вызов `ds_search` и ссылка на источник.

Глобус Z.ai у моделей GLM лучше выключить (снять «Default on» или убрать ds_search Z.ai в «Official tools»), чтобы не платить за дорогой поиск Z.ai.

## Если что-то не так

- `systemctl status deepseek-search` и `journalctl -u deepseek-search -n 30` — состояние и ошибки.
- «DeepSeek ответил 401» — неверный ключ: заново запустить блок из шага 1.
- «DeepSeek ответил 402» — нет денег на балансе DeepSeek.
- Некоторые сайты DeepSeek не читает (ds_fetch даёт 403/404) — модель ответит по выжимке из поиска.
- Удалить: `systemctl disable --now deepseek-search && rm -rf /opt/deepseek-search /etc/deepseek-search.env /etc/systemd/system/deepseek-search.service && ufw delete allow from 172.19.0.0/16 to 172.19.0.1 port 8767 proto tcp`
- Если раньше ставили сервер Z.ai (`zai-search`), его удалить так: `systemctl disable --now zai-search && rm -rf /opt/zai-search /etc/zai-search.env /etc/systemd/system/zai-search.service && ufw delete allow from 172.19.0.0/16 to 172.19.0.1 port 8766 proto tcp` (и сервер `web` в админке DEEIX).

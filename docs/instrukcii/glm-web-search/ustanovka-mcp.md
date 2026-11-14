# Веб-поиск для GLM вместе с pravo: MCP-сервер zai-search

## Почему не работало

Проверено 2026-10-25: если в запросе к Z.ai кроме встроенного поиска (`web_search`) есть обычные
инструменты (наш pravo), Z.ai в интернете не ищет, а модель вызывает pravo. Без pravo поиск работает.
Это поведение Z.ai, DEEIX тут ни при чём.

## Что делаем

Ставим на VPS маленький MCP-сервер `zai-search` (рядом с мостом pravo). Он даёт модели инструмент
`web_search`, который ходит в Z.ai Web Search API. Теперь поиск для модели такой же инструмент, как pravo,
и они работают вместе. Подойдёт и для любых других моделей (DeepSeek и т. д.).

- Работает на том же Python, что и pravo (`/opt/pravo-mcp/venv`), ничего доустанавливать не нужно.
- Слушает `172.19.0.1:8766/mcp`, только для контейнера DEEIX, наружу не открыт.
- Ключ Z.ai хранится в `/etc/zai-search.env` (доступ только root).
- Оплата: каждый поиск списывается с баланса Z.ai (Web Search API), и это считается в лимит вызовов инструментов DEEIX.

## Шаг 1. Установка на VPS

Скопировать **весь** блок ниже целиком (до последней строки `SVCEOF` включительно) и вставить в веб-терминал 1Panel.
Он спросит ключ Z.ai (не отображается при вводе). В конце должно быть `active` и `ответ моста: 406`.

```bash
clear && read -s -p "Ключ Z.ai: " K && echo && printf 'ZAI_API_KEY=%s\n' "$K" > /etc/zai-search.env && chmod 600 /etc/zai-search.env && mkdir -p /opt/zai-search && cat > /opt/zai-search/zai_search.py <<'PYEOF' && cat > /etc/systemd/system/zai-search.service <<'SVCEOF' && systemctl daemon-reload && systemctl enable --now zai-search && ufw allow from 172.19.0.0/16 to 172.19.0.1 port 8766 proto tcp && sleep 4 && systemctl is-active zai-search && curl -s -o /dev/null -w 'ответ моста: %{http_code} (406 = норма)\n' http://172.19.0.1:8766/mcp
"""MCP-сервер веб-поиска через Z.ai Web Search API для DEEIX.

Нужен потому, что Z.ai не выполняет встроенный поиск (tools: web_search), если в запросе
есть обычные инструменты (MCP pravo и др.). Этот сервер даёт модели поиск как обычный инструмент.

Запуск: ZAI_API_KEY=... python zai_search.py  (слушает ZAI_SEARCH_HOST:ZAI_SEARCH_PORT, путь /mcp)
"""

import os

import httpx
from fastmcp import FastMCP

API_URL = os.environ.get("ZAI_SEARCH_URL", "https://api.z.ai/api/paas/v4/web_search")
API_KEY = os.environ.get("ZAI_API_KEY", "")
ENGINE = os.environ.get("ZAI_SEARCH_ENGINE", "search-prime")
HOST = os.environ.get("ZAI_SEARCH_HOST", "172.19.0.1")
PORT = int(os.environ.get("ZAI_SEARCH_PORT", "8766"))

RECENCY = {"oneDay", "oneWeek", "oneMonth", "oneYear", "noLimit"}

mcp = FastMCP("zai-search")


@mcp.tool
async def web_search(
    query: str,
    count: int = 5,
    recency: str = "noLimit",
    domain: str = "",
) -> dict:
    """Поиск в интернете (Z.ai Web Search). Используй для любых свежих или внешних сведений:
    новости, курсы валют, цены, события, сведения о компаниях и людях, всё, чего нет в других
    инструментах. query — поисковый запрос на языке источников; count — сколько результатов (1–20);
    recency — свежесть: oneDay, oneWeek, oneMonth, oneYear или noLimit; domain — искать только
    на этом сайте (например cbr.ru), пусто — везде. В ответе указывай источники ссылками (link)."""
    query = (query or "").strip()
    if not query:
        return {"error": "пустой запрос"}
    if not API_KEY:
        return {"error": "на сервере не задан ZAI_API_KEY"}
    body = {
        "search_engine": ENGINE,
        "search_query": query[:300],
        "count": max(1, min(int(count or 5), 20)),
        "search_recency_filter": recency if recency in RECENCY else "noLimit",
    }
    if domain.strip():
        body["search_domain_filter"] = domain.strip()
    try:
        async with httpx.AsyncClient(timeout=40) as client:
            resp = await client.post(
                API_URL, json=body, headers={"Authorization": f"Bearer {API_KEY}"}
            )
    except httpx.HTTPError as exc:
        return {"error": f"Z.ai недоступен: {exc.__class__.__name__}"}
    try:
        data = resp.json()
    except ValueError:
        data = {}
    if resp.status_code != 200:
        err = data.get("error") if isinstance(data, dict) else None
        return {"error": f"Z.ai ответил {resp.status_code}", "details": err or resp.text[:500]}
    items = data.get("search_result") or []
    if isinstance(items, dict):
        items = [items]
    results = []
    for item in items:
        if not isinstance(item, dict):
            continue
        results.append(
            {
                "title": item.get("title", ""),
                "link": item.get("link", ""),
                "site": item.get("media", ""),
                "date": item.get("publish_date", ""),
                "content": (item.get("content") or "")[:3000],
            }
        )
    return {"query": query, "results": results}


if __name__ == "__main__":
    mcp.run(transport="http", host=HOST, port=PORT, path="/mcp")
PYEOF
[Unit]
Description=Z.ai web search MCP for DEEIX
After=network-online.target docker.service
Wants=network-online.target

[Service]
DynamicUser=yes
StateDirectory=zai-search
Environment=HOME=/var/lib/zai-search
EnvironmentFile=/etc/zai-search.env
ExecStart=/opt/pravo-mcp/venv/bin/python /opt/zai-search/zai_search.py
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVCEOF
```

## Шаг 2. Подключить в DEEIX

Админка → Инструменты → MCP → Добавить сервер:
- Имя: `web`
- URL: `http://172.19.0.1:8766/mcp`
- Статус: включён

Нажать **Sync**: должен появиться один инструмент `web_search`.

## Шаг 3. Проверка

Новый чат → GLM → включить кнопку MCP-инструментов (pravo и web) → «Какой курс доллара ЦБ РФ на сегодня? Укажи источник».
В ответе должен быть вызов `web_search` и ссылка на источник.

Встроенный поиск Z.ai (глобус) можно оставить: он сработает, когда MCP-инструменты выключены.

## Если что-то не так

- `systemctl status zai-search` и `journalctl -u zai-search -n 30` — состояние и ошибки сервера.
- Ошибка «Z.ai ответил 400» про движок: в `/etc/zai-search.env` добавить строку `ZAI_SEARCH_ENGINE=search_pro_jina`, затем `systemctl restart zai-search`.
- Ошибка 1113: на балансе Z.ai нет денег для Web Search API.
- Сменить ключ: заново запустить блок из шага 1 (файлы перезапишутся).
- Удалить: `systemctl disable --now zai-search && rm -rf /opt/zai-search /etc/zai-search.env /etc/systemd/system/zai-search.service && ufw delete allow from 172.19.0.0/16 to 172.19.0.1 port 8766 proto tcp`

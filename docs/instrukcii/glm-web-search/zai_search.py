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

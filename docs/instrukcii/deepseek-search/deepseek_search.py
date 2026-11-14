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

#!/usr/bin/env python3
"""Импорт скиллов ru-legal в DEEIX-Chat (сверено с исходниками тега v0.4.4-beta.4).

Три режима (только стандартная библиотека Python 3.9+):

  build   читает скиллы из клона ru-legal и собирает файл-набор JSON (ничего не отправляет);
  upload  загружает набор в DEEIX; БЕЗ --apply это только пробный прогон без сети;
  remove  удаляет из DEEIX скиллы с заданным префиксом триггера (тоже только с --apply).

Что делает build по ограничениям DEEIX (в символах):
  title <= 64, trigger <= 64, description <= 256, markdown <= 10000.
  YAML-шапка SKILL.md вырезается (DEEIX её не разбирает), длинные скиллы режутся
  на части по заголовкам и коду (каждая часть получает пометку «часть i из n»).
  Триггер = <пак>-<скилл>[-<номер части>] (например contract-law-nda-draft-1).

API (по исходникам): POST /api/v1/auth/login {username,password};
  общие скиллы (нужна роль админа): GET/POST /api/v1/admin/skills, PATCH/DELETE .../{id};
  личные: GET/POST /api/v1/skills/mine, PATCH/DELETE .../{id}.
  Поля: title, trigger, description, markdown, enabled, sortOrder.
"""
import argparse
import getpass
import json
import os
import re
import subprocess
import sys
import time
import urllib.error
import urllib.request

LIM_TITLE, LIM_TRIGGER, LIM_DESC, LIM_MD = 64, 64, 256, 10000
SERVICE_SKILLS = {"cold-start-interview", "customize"}  # работают с файлами профиля на диске

FIRST_NOTE = (
    "[Источник: ru-legal, пак {pack}, Apache-2.0, коммит {commit}. Среда: чат DEEIX без доступа "
    "к файлам и командам. Если скилл просит прочитать профиль или файл, спроси нужные данные у "
    "пользователя. Инструменты (например pravo) доступны, только если выбраны в сообщении.]"
)
PART_NOTE = "[{key} · часть {i} из {n} — подключай все части вместе]"


# ---------------------------------------------------------------- разбор SKILL.md
def parse_frontmatter(text):
    m = re.match(r"^---[ \t]*\r?\n(.*?)\r?\n---[ \t]*(?:\r?\n|$)", text, re.S)
    if not m:
        return {}, text
    data, lines, i = {}, m.group(1).split("\n"), 0
    while i < len(lines):
        line = lines[i]
        mk = re.match(r"^([A-Za-z_][\w-]*):[ \t]*(.*)$", line)
        if mk and not line.startswith((" ", "\t")):
            key, val = mk.group(1), mk.group(2).strip()
            if val in (">", "|", ">-", "|-"):
                block = []
                i += 1
                while i < len(lines) and (lines[i].startswith((" ", "\t")) or not lines[i].strip()):
                    block.append(lines[i].strip())
                    i += 1
                joiner = " " if val.startswith(">") else "\n"
                data[key] = joiner.join(x for x in block if x)
                continue
            if val[:1] in ('"', "'"):
                end = val.find(val[0], 1)
                val = val[1:end] if end > 0 else val.strip("\"'")
            else:
                val = re.sub(r"\s+#.*$", "", val)
            data[key] = val
        i += 1
    return data, text[m.end():]


def shorten(text, limit):
    text = re.sub(r"\s+", " ", text).strip()
    if len(text) <= limit:
        return text
    cut = text[: limit - 1]
    sentence = max(cut.rfind(". "), cut.rfind("; "))
    if sentence > limit * 0.5:
        return cut[: sentence + 1].rstrip()
    return cut[: cut.rfind(" ")].rstrip(" ,;:") + "…"


# ---------------------------------------------------------------- нарезка длинного текста
def _sections(body):
    secs, cur, fence = [], [], False
    for line in body.split("\n"):
        if line.lstrip().startswith("```"):
            fence = not fence
        elif not fence and re.match(r"^#{1,2}\s", line) and cur:
            secs.append("\n".join(cur))
            cur = []
        cur.append(line)
    if cur:
        secs.append("\n".join(cur))
    return secs


def _units(text, limit):
    """Блоки по пустым строкам вне кода; слишком длинные режутся по строкам."""
    units, cur, fence = [], [], False
    for line in text.split("\n"):
        if line.lstrip().startswith("```"):
            fence = not fence
        if not line.strip() and not fence and cur:
            units.append("\n".join(cur))
            cur = []
            continue
        cur.append(line)
    if cur:
        units.append("\n".join(cur))
    out = []
    for u in units:
        if len(u) <= limit:
            out.append(u)
            continue
        piece = ""
        for ln in u.split("\n"):
            while len(ln) > limit:  # одна строка длиннее лимита
                out.append(ln[:limit])
                ln = ln[limit:]
            if piece and len(piece) + 1 + len(ln) > limit:
                out.append(piece)
                piece = ln
            else:
                piece = f"{piece}\n{ln}" if piece else ln
        if piece:
            out.append(piece)
    return out


def split_markdown(body, limit):
    chunks, cur = [], ""

    def add(piece, sep):
        nonlocal cur
        if cur and len(cur) + len(sep) + len(piece) > limit:
            chunks.append(cur)
            cur = ""
        cur = f"{cur}{sep}{piece}" if cur else piece

    for sec in _sections(body.strip("\n")):
        if len(sec) <= limit:
            add(sec, "\n")
        else:
            for unit in _units(sec, limit):
                add(unit, "\n\n")
    if cur:
        chunks.append(cur)
    return [c.strip() for c in chunks if c.strip()]


def make_parts(pack, name, desc, body, commit):
    key = f"{pack}/{name}"
    first_note = FIRST_NOTE.format(pack=pack, commit=commit)
    single = f"{first_note}\n\n{body.strip()}"
    if len(single) <= LIM_MD:
        return [dict(trigger=f"{pack}-{name}", title=key, description=desc, markdown=single)]
    reserve = len(first_note) + 2 + len(PART_NOTE.format(key=key, i=99, n=99)) + 2 + 60
    limit = LIM_MD - reserve
    while limit > 1000:
        chunks = split_markdown(body, limit)
        n = len(chunks)
        parts = []
        for i, chunk in enumerate(chunks, 1):
            head = PART_NOTE.format(key=key, i=i, n=n)
            text = f"{first_note}\n\n{head}\n\n{chunk}" if i == 1 else f"{head}\n\n{chunk}"
            parts.append(
                dict(
                    trigger=f"{pack}-{name}-{i}",
                    title=f"{key} ({i}/{n})",
                    description=desc if i == 1 else f"Продолжение скилла {key} (часть {i} из {n}). Подключай вместе с частью 1.",
                    markdown=text,
                )
            )
        if all(len(p["markdown"]) <= LIM_MD for p in parts):
            return parts
        limit -= 300
    raise SystemExit(f"не удалось нарезать скилл {key} в лимит {LIM_MD}")


# ---------------------------------------------------------------- build
def cmd_build(a):
    root = os.path.join(a.repo, "packs", a.pack, "skills")
    if not os.path.isdir(root):
        raise SystemExit(f"нет каталога {root}")
    try:
        commit = subprocess.run(
            ["git", "-C", a.repo, "rev-parse", "--short", "HEAD"], capture_output=True, text=True, check=True
        ).stdout.strip()
    except Exception:
        commit = "unknown"
    skills, skipped, order = [], [], 0
    for d in sorted(os.listdir(root)):
        path = os.path.join(root, d, "SKILL.md")
        if not os.path.isfile(path):
            continue
        fm, body = parse_frontmatter(open(path, encoding="utf-8").read())
        name = fm.get("name") or d
        if name in SERVICE_SKILLS and not a.include_service:
            skipped.append({"skill": f"{a.pack}/{name}", "reason": "служебный: работает с файлами профиля на диске"})
            continue
        parts = make_parts(a.pack, name, shorten(fm.get("description", ""), LIM_DESC), body, commit)
        for p in parts:
            order += 1
            p["sortOrder"] = order
            for field, lim in (("title", LIM_TITLE), ("trigger", LIM_TRIGGER), ("description", LIM_DESC), ("markdown", LIM_MD)):
                if not p[field] or len(p[field]) > lim:
                    raise SystemExit(f"{p['trigger']}: поле {field} вне лимита ({len(p[field])} > {lim})")
        skills.append({"key": f"{a.pack}/{name}", "parts": parts})
    data = {"source": {"pack": a.pack, "commit": commit}, "skills": skills, "skipped": skipped}
    with open(a.out, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=1)
    print_plan(data)
    print(f"\nнабор записан: {a.out} (коммит {commit})")


def print_plan(data, only=None):
    total_parts = total_chars = 0
    print(f"{'скилл':34} {'частей':>6} {'символов':>9}  триггеры")
    for s in data["skills"]:
        parts = [p for p in s["parts"] if not only or only in p["trigger"] or only in s["key"]]
        if not parts:
            continue
        chars = sum(len(p["markdown"]) for p in parts)
        total_parts += len(parts)
        total_chars += chars
        trig = parts[0]["trigger"] if len(parts) == 1 else f'{parts[0]["trigger"]} … {parts[-1]["trigger"]}'
        print(f"{s['key']:34} {len(parts):>6} {chars:>9}  /{trig}")
    print(f"{'ИТОГО':34} {total_parts:>6} {total_chars:>9}")
    for sk in data.get("skipped", []):
        print(f"пропущен: {sk['skill']} — {sk['reason']}")


# ---------------------------------------------------------------- HTTP
class Api:
    def __init__(self, base):
        self.base = base.rstrip("/")
        self.token = None

    def call(self, method, path, body=None, query=None):
        url = f"{self.base}/api/v1{path}"
        if query:
            url += "?" + "&".join(f"{k}={v}" for k, v in query.items())
        data = json.dumps(body, ensure_ascii=False).encode("utf-8") if body is not None else None
        req = urllib.request.Request(url, data=data, method=method)
        req.add_header("Content-Type", "application/json")
        if self.token:
            req.add_header("Authorization", f"Bearer {self.token}")
        try:
            with urllib.request.urlopen(req, timeout=60) as r:
                raw, status = r.read(), r.status
        except urllib.error.HTTPError as e:
            raw, status = e.read(), e.code
        except urllib.error.URLError as e:
            raise SystemExit(f"нет связи с {self.base}: {e.reason}")
        try:
            env = json.loads(raw.decode("utf-8"))
        except Exception:
            env = {"errorMsg": raw[:200].decode("utf-8", "replace"), "data": None}
        return status, env


def login(api, username):
    password = os.environ.get("DEEIX_PASSWORD") or getpass.getpass(f"Пароль для {username}: ")
    st, env = api.call("POST", "/auth/login", {"username": username, "password": password})
    if st != 200:
        raise SystemExit(f"вход не удался (HTTP {st}): {env.get('errorMsg') or env.get('errorCode')}")
    d = env["data"]
    if d.get("twoFactorRequired"):
        methods = d.get("verificationMethods") or []
        method = "two_factor" if "two_factor" in methods or not methods else "email"
        challenge = d.get("twoFactorChallengeToken")
        if method == "email":
            api.call("POST", "/auth/2fa/email/start", {"challengeToken": challenge})
            print("Код отправлен на почту.")
        code = input("Код двухфакторной проверки: ").strip()
        st, env = api.call(
            "POST", "/auth/2fa/verify", {"challengeToken": challenge, "verificationMethod": method, "code": code}
        )
        if st != 200:
            raise SystemExit(f"код не принят (HTTP {st}): {env.get('errorMsg') or env.get('errorCode')}")
        d = env["data"]
    if not d.get("accessToken"):
        raise SystemExit("в ответе входа нет accessToken")
    api.token = d["accessToken"]


def base_path(mine):
    return "/skills/mine" if mine else "/admin/skills"


def list_existing(api, mine):
    st, env = api.call("GET", base_path(mine), query={"page": 1, "page_size": 1000})
    explain_failure(st, env, mine, "список скиллов")
    return {s["trigger"]: s["id"] for s in env["data"]["results"]}


def explain_failure(st, env, mine, what):
    if st == 200:
        return
    hint = ""
    if st in (401, 403) and not mine:
        hint = " Нужна роль администратора; для личных скиллов запусти с --mine."
    raise SystemExit(f"{what}: HTTP {st}: {env.get('errorMsg') or env.get('errorCode')}.{hint}")


def connect(a):
    api = Api(a.base)
    login(api, a.user)
    return api


# ---------------------------------------------------------------- upload / remove
def cmd_upload(a):
    data = json.load(open(a.file, encoding="utf-8"))
    print_plan(data, a.only)
    parts = [p for s in data["skills"] for p in s["parts"] if not a.only or a.only in p["trigger"] or a.only in s["key"]]
    if not parts:
        raise SystemExit("по фильтру --only ничего не найдено")
    if not a.apply:
        print("\nПРОБНЫЙ ПРОГОН: ничего не отправлено. Для загрузки добавь --apply.")
        return
    api = connect(a)
    existing = list_existing(api, a.mine)
    created = updated = 0
    for p in parts:
        payload = {k: p[k] for k in ("title", "trigger", "description", "markdown", "sortOrder")}
        payload["enabled"] = True
        if p["trigger"] in existing:
            st, env = api.call("PATCH", f"{base_path(a.mine)}/{existing[p['trigger']]}", payload)
            verb = "обновлён"
        else:
            st, env = api.call("POST", base_path(a.mine), payload)
            verb = "создан"
        if st != 200:
            raise SystemExit(f"/{p['trigger']}: HTTP {st}: {env.get('errorMsg') or env.get('errorCode')}")
        updated += verb == "обновлён"
        created += verb == "создан"
        print(f"  {verb}: /{p['trigger']}  (id {env['data']['skill']['id']}, {len(p['markdown'])} симв.)")
        time.sleep(0.2)
    print(f"\nготово: создано {created}, обновлено {updated}")


def cmd_remove(a):
    api = connect(a)
    existing = list_existing(api, a.mine)
    victims = {t: i for t, i in existing.items() if t.startswith(a.prefix)}
    if not victims:
        print(f"скиллов с префиксом «{a.prefix}» нет")
        return
    for t in sorted(victims):
        print(f"  к удалению: /{t} (id {victims[t]})")
    if not a.apply:
        print("\nПРОБНЫЙ ПРОГОН: ничего не удалено. Для удаления добавь --apply.")
        return
    for t, i in sorted(victims.items()):
        st, env = api.call("DELETE", f"{base_path(a.mine)}/{i}")
        if st != 200:
            raise SystemExit(f"/{t}: HTTP {st}: {env.get('errorMsg') or env.get('errorCode')}")
        print(f"  удалён: /{t}")
    print(f"\nудалено: {len(victims)}")


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd", required=True)

    b = sub.add_parser("build", help="собрать набор из клона ru-legal")
    b.add_argument("--repo", required=True, help="путь к клону ru-legal")
    b.add_argument("--pack", default="contract-law")
    b.add_argument("--out", required=True)
    b.add_argument("--include-service", action="store_true", help="не пропускать cold-start-interview и customize")
    b.set_defaults(fn=cmd_build)

    def net(p):
        p.add_argument("--base", default="http://127.0.0.1:8080", help="адрес DEEIX (по умолчанию локальный)")
        p.add_argument("--user", required=True, help="имя пользователя DEEIX")
        p.add_argument("--mine", action="store_true", help="личные скиллы вместо общих (не нужна роль админа)")
        p.add_argument("--apply", action="store_true", help="выполнить; без флага — пробный прогон")

    u = sub.add_parser("upload", help="загрузить набор в DEEIX")
    u.add_argument("file")
    u.add_argument("--only", help="часть триггера или имя скилла, например nda-draft")
    net(u)
    u.set_defaults(fn=cmd_upload)

    r = sub.add_parser("remove", help="удалить скиллы по префиксу триггера")
    r.add_argument("--prefix", required=True, help="например contract-law-")
    net(r)
    r.set_defaults(fn=cmd_remove)

    a = ap.parse_args()
    a.fn(a)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        sys.exit("прервано")

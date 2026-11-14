# Документация DEEIX-RU

Главные правила для агента — в [`../AGENTS.md`](../AGENTS.md).

## Основное

- [vps.md](vps.md) — сервер: пути, контейнеры, сервисы, порты, бэкап.
- [sborka.md](sborka.md) — как устроен аддон, как вносить правку, проверка, сборка, обновление DEEIX.
- [zametki/preferences.md](zametki/preferences.md), [zametki/deeix-ru.md](zametki/deeix-ru.md) — исходные договорённости проекта.

## Правки DEEIX (все слиты и стоят на VPS, если не сказано иначе)

| Что | Патч / папка | PR | Заметка | Инструкция |
|---|---|---|---|---|
| Русский язык, рубли, автоназвания, подсветка «мозга» | `ru-locale` | — , #11 | [brain-highlight](zametki/deeix-brain-highlight.md) | — |
| PDF-сканы без текста → картинки страниц | `scan-pdf` | #1, #2 | [vps-memory-scan-pdf](zametki/vps-memory-scan-pdf.md) | — |
| DeepSeek по протоколу Anthropic + инструменты | `deepseek-anthropic` | #3 | [deepseek-anthropic](zametki/deeix-deepseek-anthropic.md) | — |
| Иконка «D» | `app-icon` | #4 | [app-icon](zametki/deeix-app-icon.md) | — |
| Переход на v0.4.5 / v0.4.6 | — | #5, #18 | [v045](zametki/deeix-upgrade-v045.md), [v046](zametki/deeix-upgrade-v046.md) | [v0.4.5](instrukcii/deeix-update/obnovlenie-v0.4.5.md), [v0.4.6](instrukcii/deeix-update/obnovlenie-v0.4.6.md) |
| Широкие таблицы | `table-fit` | #6 | [table-fit](zametki/deeix-table-fit.md) | [ustanovka](instrukcii/table-fit/ustanovka.md) |
| HTML-вёрстка в ответе | `inline-html` | #7 | [html-visual](zametki/deeix-html-visual.md) | [ustanovka](instrukcii/inline-html/ustanovka.md) |
| «/» и «@» в меню «+» | `composer-triggers` | #8, #12 | [composer-triggers](zametki/deeix-composer-triggers.md) | [ustanovka](instrukcii/composer-triggers/ustanovka.md) |
| Экспорт в Word со шрифтами с экрана | `files/.../export-docx.ts` | #9 | [docx-export](zametki/deeix-docx-export.md) | [ustanovka](instrukcii/docx-fonts/ustanovka.md) |
| Строка «Скиллы / Инструменты» под ответом | `turn-resources` | #10 | [usage-line](zametki/deeix-usage-line.md) | [ustanovka](instrukcii/turn-resources/ustanovka.md) |
| Поле ввода внизу на пустом экране | `composer-bottom` | #13 | [composer-bottom](zametki/deeix-composer-bottom.md) | [ustanovka](instrukcii/composer-bottom/ustanovka.md) |
| Перевод строки трассы про скиллы | `ru-locale` | #14 | [tools-not-called](zametki/deeix-tools-not-called.md) | [ustanovka](instrukcii/skill-trace-ru/ustanovka.md) |
| Китайские тексты сервера + статистика на SQLite | `ru-locale`, `usage-stats` | #15 | [server-text-ru](zametki/deeix-server-text-ru.md) | [ustanovka](instrukcii/server-text-ru/ustanovka.md) |
| Кнопка ↕ (скрыть заголовок и поле ввода) | `focus-toggle` | #16 | [focus-toggle](zametki/deeix-focus-toggle.md) | [ustanovka](instrukcii/focus-toggle/ustanovka.md) |
| Чтение `.doc` | `doc-text` | #17 | [doc-text](zametki/deeix-doc-text.md) | [ustanovka](instrukcii/doc-text/ustanovka.md) |
| Своё OCR (Tesseract) для текстовых моделей | `ocr-text`, `ocr-service/` | #19 | [own-ocr](zametki/deeix-own-ocr.md) | [ustanovka](instrukcii/own-ocr/ustanovka.md) |
| Ширина боковой панели мышкой | `sidebar-width` | #20 | [sidebar-width](zametki/deeix-sidebar-width.md) | [ustanovka](instrukcii/sidebar-width/ustanovka.md) |
| Компактное меню боковой панели | `sidebar-menu` | #21 | [sidebar-menu](zametki/deeix-sidebar-menu.md) | [ustanovka](instrukcii/sidebar-menu/ustanovka.md) |
| Прокрутка вложений | `attachment-scroll` | #22 | [attachment-scroll](zametki/deeix-attachment-scroll.md) | [ustanovka](instrukcii/attachment-scroll/ustanovka.md) |
| Правки и примечания Word для модели | `word-revisions` | #23 | [word-revisions](zametki/deeix-word-revisions.md) | [ustanovka](instrukcii/word-revisions/ustanovka.md) |
| Таблицы-реестры, «во всю ширину» | `data-table-fit` | #24 | [data-table-fit](zametki/deeix-data-table-fit.md) | [ustanovka](instrukcii/data-table-fit/ustanovka.md) |
| Уровень рассуждения и стоимость чата | `effort-cost` | #25 | [effort-cost](zametki/deeix-effort-cost.md) | [ustanovka](instrukcii/effort-cost/ustanovka.md) |
| «(2)» — число веток в названии чата | `branch-count` | #26 | [branch-count](zametki/deeix-branch-count.md) | [ustanovka](instrukcii/branch-count/ustanovka.md) |
| Маскировка ПД (прокси) | `pd-mask/` | #27 | [pd-masking](zametki/deeix-pd-masking.md) | [ustanovka](instrukcii/pd-mask/ustanovka.md) |
| Окно «Что увидит модель» | `pd-preview` | #28 | [pd-masking](zametki/deeix-pd-masking.md) | [ustanovka](instrukcii/pd-preview/ustanovka.md) |

Возможная доработка (Kimi упоминал): список исключений ФИО для маскировки ПД.

## Модели, поиск, MCP

- [zametki/deeix-model-catalog.md](zametki/deeix-model-catalog.md) — почему модель не видна в чате; цены кэша.
- [zametki/deeix-glm-flashx-legal.md](zametki/deeix-glm-flashx-legal.md) — JSON возможностей GLM.
- [zametki/deeix-deepseek-search.md](zametki/deeix-deepseek-search.md) — веб-поиск DeepSeek; MCP-сервер [deepseek_search.py](instrukcii/deepseek-search/deepseek_search.py), [установка](instrukcii/deepseek-search/ustanovka.md).
- [zametki/deeix-glm-web-search.md](zametki/deeix-glm-web-search.md) — поиск GLM (Z.ai); старый `zai_search.py` больше не нужен.
- [zametki/deeix-tools-not-called.md](zametki/deeix-tools-not-called.md) — модель не вызывает инструменты: где промпты-тормоза, как проверить по БД.
- [zametki/deeix-masking-timeout.md](zametki/deeix-masking-timeout.md) — 503 masking_timeout у агрегатора ([схема](instrukcii/masking-timeout/skhema.html)).
- [zametki/deeix-pro-hang.md](zametki/deeix-pro-hang.md) — «зависания» deepseek-v4-pro; готовые проверки в [instrukcii/deepseek-pro-hang/](instrukcii/deepseek-pro-hang/).
- pravo-mcp (MCP pravo.gov.ru) — подробно в [arhiv/STATE-2026-10-15.md](arhiv/STATE-2026-10-15.md), часть B.
- [zametki/egrul-kad.md](zametki/egrul-kad.md) — ЕГРЮЛ/КАД: **не ставить** (решение Kimi 2026-11-14).

## Скиллы для DEEIX

- [zametki/deeix-skills-ru.md](zametki/deeix-skills-ru.md) — русский набор по договорному праву (`/договор-…`): [dogovor-skills.json](instrukcii/skills-ru/dogovor-skills.json), [import_skills.py](instrukcii/skills-ru/import_skills.py), тексты в [instrukcii/skills-ru/texts/](instrukcii/skills-ru/texts/), откат в `old/`.
- [zametki/ru-legal-audit.md](zametki/ru-legal-audit.md) — проверка ru-legal на соответствие праву РФ ([отчёт](instrukcii/ru-legal-audit/ru-legal-proverka.md)).
- [zametki/deeix-chat-skills.md](zametki/deeix-chat-skills.md) — одиночные скиллы: [eli5](instrukcii/skills-eli5/eli5-deeix.md), [brainstorming](instrukcii/skills-brainstorming/brainstorming-deeix.md), [критика науки](instrukcii/skills-science/kritika-nauki-deeix.md).

## Разное

- [zametki/deeix-sharing.md](zametki/deeix-sharing.md) — публичные ссылки на диалог.
- [zametki/deeix-db-backup.md](zametki/deeix-db-backup.md), [instrukcii/db-backup/kopiya-bd.md](instrukcii/db-backup/kopiya-bd.md) — копия базы.
- `zametki/vps-cve-*.md` — разборы уязвимостей для VPS.
- Android-оболочка: [zametki/tap-shell-app.md](zametki/tap-shell-app.md), [zametki/tap-shell-automerge.md](zametki/tap-shell-automerge.md), [zametki/apk-versioning.md](zametki/apk-versioning.md) (код — в репозитории tap-shell).

## Архив

- [arhiv/STATE-2026-10-15.md](arhiv/STATE-2026-10-15.md) — первая большая сводка (октябрь, версия v0.4.4-beta.4): устройство скиллов и MCP в DEEIX, pravo-mcp, принятые решения, грабли. Часть сведений устарела — свежее в `zametki/`.
- [arhiv/HANDOFF-2026-10-16.md](arhiv/HANDOFF-2026-10-16.md), [arhiv/SCAN-PDF.md](arhiv/SCAN-PDF.md) — разбор зависания VPS на PDF-сканах.

Скриншоты к инструкциям в репозиторий не кладутся (они были только для показа Kimi).

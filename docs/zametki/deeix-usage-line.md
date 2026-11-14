# deeix-usage-line

_Строка «Скиллы / Инструменты» под ответом DEEIX — patches/turn-resources.patch, PR #10 (2026-10-25), откуда данные._

- 2026-10-25 по просьбе Kimi: под строкой модель/токены/цена (message-meta.tsx, AssistantMessageMeta) вторая строка «Скиллы: …» и «Инструменты: name ×N». Парсер features/chat/model/message-turn-resources.ts.
- Данные из processTrace без правок сервера (v0.4.5): скиллы — payload process-блока skill_titles/skill_triggers (запасной путь promptTrace.blocks skill_context sourceRefs); инструменты — tool_calls[].name в событиях stage=tool, дедуп по tool_call_id, блок tools только если в событиях пусто. Нужен chat.process_trace_enabled (по умолчанию true). Имена MCP в журнале без имени сервера (search_npa, не pravo).
- Скиллы = подключены к запросу (использовала ли модель — неизвестно); инструменты = реальные вызовы.
- Патч patches/turn-resources.patch (флаг --no-turn-resources), ключи chat.meta.skillsUsed/toolsUsed/…Hint/toolsFailedCalls (en/zh в патче, ru в files/). PR https://github.com/Dyslexia2813/deeix-ru-build/pull/10, инструкция ../instrukcii/turn-resources/ustanovka.md (копия out-pered-stroka).
- 2026-10-25 в ветку PR #10 влит PR #9 (docx-шрифты), общая сборка run 40226089437. Kimi проверил на VPS («работает»), PR #10 слит 2026-10-25, #9 закрылся как слитый. Стоит на сервере.
- Локальная проверка: ответ с журналом вставляется прямо в sqlite (chat_conversations, chat_messages, chat_runs, chat_run_events с event_scope trace_block/trace_event); онбординг — identity_users.onboarding_completed_at.

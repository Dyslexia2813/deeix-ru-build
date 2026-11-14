# deeix-deepseek-anthropic

_DeepSeek в DEEIX по протоколу Anthropic ломает вызовы инструментов (400 thinking, сырой DSML): причина, обход настройкой, PR #3._

- 2026-10-17 модель deepseek-anthropic-WEB с инструментом pravo: с рассуждением 400 «content[].thinking ... must be passed back», без рассуждения ответ = сырой DSML-текст вызова.
- Причина (DEEIX v0.4.4-beta.4): передача рассуждения и разбор DSML есть только для протокола OpenAI Chat Completions (service_view_normalization.go reasoningContentPassbackRequired, infra/llm/dsml_tool_calls.go); адаптер Anthropic thinking-блоки не возвращает и DSML не разбирает.
- Обход без сборки: подключить DeepSeek по протоколу OpenAI Chat Completions, адрес https://api.deepseek.com.
- Исправление в коде: patches/deepseek-anthropic.patch, PR https://github.com/Dyslexia2813/deeix-ru-build/pull/3 (флаг install.sh --no-deepseek-anthropic). На живом DeepSeek не проверено. См. [deeix-model-catalog](deeix-model-catalog.md), [vps-memory-scan-pdf](vps-memory-scan-pdf.md).
- 2026-10-18 бинарник из ветки PR #3 (run 39112703223, включает PR #2) установлен на VPS (копия deeix-chat.prev); с рассуждением pravo работает. PR #3 ещё не слит.
- Ошибка «Выполнение инструментов достигло лимита»: лимиты DEEIX на сообщение, по умолчанию 8 MCP-вызовов и 5 запросов к LLM (Админка → Инструменты). Предложено 20 и 8.
- 2026-10-18 вторая ошибка: при исчерпании запросов к LLM итоговый шаг без инструментов, DeepSeek продолжал DSML → final_answer_missing. Исправлено коммитом 1ca557a в PR #3 (история инструментов текстом на этом шаге, подпись thinking сохраняется в normalizeMessages). Сборка run 39122044858 поставлена на VPS 2026-10-18, Kimi подтвердил: работает. PR #3 слит 2026-10-18 (e193822).

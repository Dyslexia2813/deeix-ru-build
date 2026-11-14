# deeix-effort-cost

_Уровень рассуждения рядом с моделью под ответом и стоимость всего чата в шапке — patches/effort-cost.patch, PR #25 (сайт + бэкенд)._

- 2026-11-09 Kimi попросил: показать уровень усилий («мозг») рядом с названием модели под ответом и суммарную стоимость ответов чата слева от «Поделиться/экспорт».
- Уровень: сервер v0.4.6 уже хранит reasoning_effort в chat_messages (MessageDTO.reasoningEffort); патч только показывает его в ModelBadge (message-meta.tsx), подписи из common.reasoningEffort.levels.
- Стоимость: чат грузит только последние 30 сообщений, поэтому бэкенд добавляет billedNanousd (SUM billed_nanousd по всем сообщениям, все ветки) в GET /conversations/{id}; хук features/chat/hooks/use-chat-conversation-cost.ts, значок chat-cost-badge.tsx. Видно только при showBillingCost (биллинг не «self» + настройка чата).
- patches/effort-cost.patch (ключ --no-effort-cost, ставится последним), ключи chat.meta.reasoningEffort/conversationCost/conversationCostHint. PR https://github.com/Dyslexia2813/deeix-ru-build/pull/25 (Kimi проверил на VPS, слит 2026-11-09, squash 10ac5e6, стоит на сервере), сборка run 42677575533, инструкция ../instrukcii/effort-cost/ustanovka.md (копии out-pered-25, deeix-chat-pered-25).
- Локально: режим биллинга PATCH /admin/billing/config {"mode":"usage"}; язык профиля PATCH /me {"locale":"ru-RU"}. См. [deeix-usage-line](deeix-usage-line.md), [deeix-data-table-fit](deeix-data-table-fit.md).

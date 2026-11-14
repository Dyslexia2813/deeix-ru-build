# deeix-branch-count

_«(2)» перед названием чата в боковой панели по числу веток (перегенерации/правки) — patches/branch-count.patch, PR #26 (сайт + бэкенд)._

- 2026-11-09 Kimi попросил: когда в чате появилась 2-я ветка (перегенерированный ответ, стрелки 1/2), показывать «(2)» в начале названия, дальше (3)…
- Решение: название в базе не меняется, номер рисуется отдельным span в sidebar-conversation-item.tsx (Недавние/Избранное/Проекты). Число веток = концы дерева сообщений (сообщения без живых детей), считает бэкенд Repo.CountConversationBranches (NOT EXISTS, SQLite+Postgres), поле branchCount в списке чатов и GET /conversations/{id}.
- Живое обновление: хук features/chat/hooks/use-chat-branch-count-sync.ts запрашивает GET при открытии чата и при смене числа концов среди загруженных сообщений → touchByPublicID. preserveKnownShareState сохраняет branchCount, когда ответы rename/star его не несут.
- Ключ --no-branch-count, ставится после effort-cost; подсказка recent.row.branchCount. PR https://github.com/Dyslexia2813/deeix-ru-build/pull/26 (Kimi проверил на VPS, слит 2026-11-13, squash 989e8a6, стоит на сервере), сборка run 42702146329, инструкция ../instrukcii/branch-count/ustanovka.md (копии out-pered-26, deeix-chat-pered-26).
- Локальная проверка: модалку первичной настройки админа проще спрятать CSS в Playwright, чем проходить. См. [deeix-effort-cost](deeix-effort-cost.md).

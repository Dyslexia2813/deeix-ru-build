# deeix-brain-highlight

_Подсветка значка «мозг» (рассуждение) в поле ввода DEEIX — правило в ru-locale.patch, фикс для GLM в PR #11._

- Правило подсветки значков параметров живёт в patches/ru-locale.patch (apps/web/features/chat/components/sections/chat-model-controls.tsx, isItemActive): список с вариантом «выкл.» (none/off/disabled/false) горит, пока не «выкл.», даже на значении по умолчанию. Апстрим подсвечивает только изменённое значение.
- 2026-10-25 у GLM (Z.ai) уровни только Низкий/Высокий/Максимальный, без «выкл.» → мозг не горел на «Максимальный по умолчанию». Фикс: рассуждение без «выкл.» горит на любом уровне (кроме auto). PR #11 в deeix-ru-build, ветка fix/glm-brain-highlight, собран вместе с #1–#10; слит 2026-10-26.
- Отдельного «brain»-патча нет, ищи в ru-locale.patch. См. [deeix-glm-web-search](deeix-glm-web-search.md), [deeix-usage-line](deeix-usage-line.md).

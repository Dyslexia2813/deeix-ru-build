# deeix-md-table-full

_Обычные (markdown) таблицы в ответе: перенос длинных заголовков и кнопка «Открыть во всю ширину» — patches/md-table-full.patch._

- 2026-11-25 Kimi: у обычной таблицы нет кнопки «Открыть во всю ширину» (она была только у реестров data-table, [deeix-data-table-fit](deeix-data-table-fit.md)), а длинные заголовки столбцов наезжают друг на друга.
- Причина наезда: Streamdown (2.7.0, `MarkdownTh`) даёт `th` утилиту `whitespace-nowrap`. Слой utilities в Tailwind 4 сильнее слоя components, куда импортирован `markdown-table.css`, поэтому `white-space: normal` у `.markdown-table-cell--normal` на заголовок не действовал. После [deeix-table-fit](deeix-table-fit.md) ширина текстовых столбцов ограничена (normal ≤13.75rem), и нерастяжимый заголовок вылезал в соседний столбец. Лечится `white-space: normal !important` у `.markdown-table th` (important из более раннего слоя бьёт обычные правила позднего).
- Кнопка: `adaptive-markdown-table.tsx` — под готовой таблицей (не во время потока) «Открыть во всю ширину», с мышкой видна при наведении (`group/md-table`), как у реестров. Dialog на всё окно: название (`findTableCaption`, иначе «Таблица»), меню скачивания, крестик. Класс `markdown-table-fullscreen`: normal ≤22rem, content ≤44rem; на телефоне (<30rem) min 11rem и прокрутка вбок.
- Патч ставится после `quote-select`, нужен `table-fit` и `table-xlsx` (правит те же файлы). Ключ `--no-md-table-full`. Только сайт (`out`), строка для проверки: `table-fullscreen-action`.
- В `pnpm --filter @deeix/web check` есть старые ошибки не из этого патча (chat-cost-badge.tsx от effort-cost — biome a11y; chat-pd-preview.tsx и use-chat-conversation-cost.ts — check:arch). typecheck чистый.

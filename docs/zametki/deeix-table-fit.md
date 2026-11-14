# deeix-table-fit

_Широкие таблицы в ответах DEEIX — patches/table-fit.patch (CSS markdown-table.css), PR #6, текст инструкции для модели._

- 2026-10-22 Kimi пожаловался: таблицы в ответах уходят за край, нужна прокрутка вбок. Запрещать таблицы Kimi не хочет («таблица это удобно»), хочет уже.
- Фронтенд: apps/web/shared/components/markdown/adaptive-markdown-table.tsx (классифицирует столбцы compact/normal/content/code) + apps/web/styles/markdown-table.css (в оригинале width: max-content, min-width текстовых столбцов 12–17rem).
- patches/table-fit.patch (флаг install.sh --no-table-fit): ширина 100%, перенос по словам (break-word; anywhere рвёт слова посередине — не использовать), минимумы 8–10rem, отступ 1.25rem; на экране <30rem минимумы прежние. Нужна только замена out. PR https://github.com/Dyslexia2813/deeix-ru-build/pull/6 (слит 2026-10-22).
- 2026-10-22 вечером Kimi поставил на VPS (только out), подтвердил «всё ок»; дана команда очистки /opt/deeix-ru (оставить out, deeix-chat, compose.override.yml; удалены копии *-0.4.4, *.prev, out-pered-tablicami, zip, data-0.4.4).
- Дополнительно предлагалась инструкция модели (≤3–4 столбца, ≤5–7 слов в ячейке); где в DEEIX общий системный промпт — не проверено.
- Сборка: репо = tarball deeix-ru-addon.tar.gz + workflow build.yml (workflow_dispatch, tag). См. [deeix-upgrade-v045](deeix-upgrade-v045.md), [deeix-app-icon](deeix-app-icon.md).

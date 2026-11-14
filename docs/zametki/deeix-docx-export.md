# deeix-docx-export

_Экспорт ответа/диалога в Word (DOCX) в DEEIX-RU — свой конвертер в аддоне, что он теряет (шрифты, цвета, вёрстка), вопрос Kimi 2026-10-24._

- Конвертер: files/apps/web/shared/lib/export-docx.ts (+ zip-writer.ts) в deeix-ru-addon.tar.gz репо Dyslexia2813/deeix-ru-build; без npm-зависимостей (решение Kimi). Переводит DOM ответа по тегам в WordprocessingML.
- Что сохраняет: заголовки, жирный/курсив/зачёркнутый, sup/sub, код (Consolas), ссылки, списки, таблицы, цитаты, формулы (TeX-текстом), картинки/svg/canvas.
- Что терял до PR #9: шрифт всегда Calibri 11 (жёстко в stylesheet), не читает inline style и computed style (цвет, фон, размер, шрифт), не обрабатывает <u> и <mark>, HTML-вёрстку (карточки из [deeix-html-visual](deeix-html-visual.md)) сплющивает в абзацы.
- PDF-экспорт идёт через печать браузера и сохраняет вид как на экране.
- 2026-10-25 Kimi выбрал «шрифты с экрана»: PR https://github.com/Dyslexia2813/deeix-ru-build/pull/9 (слит 2026-10-25 вместе с PR #10, стоит на VPS, Kimi подтвердил; ветка feat/docx-screen-fonts, сборка run 40180384706). Конвертер читает getComputedStyle: шрифт = первый Word-безопасный из CSS-набора (Segoe UI, Arial, Georgia, Palatino Linotype, Consolas…), размер, цвет (любой CSS-цвет через canvas), u/маркер, фон ячеек; блоки с фоном/рамкой → абзацы с заливкой и рамкой. В тёмной теме цвета не переносятся. Инструкция ../instrukcii/docx-fonts/ustanovka.md (копия out-pered-docx). Проверка: тестовая страница в Playwright → buildDocx → LibreOffice PNG.
- 2026-10-24 Kimi спросил, можно ли сохранять шрифты и форматирование; просил пока только ответ, без переделок. Предложены варианты (стили из getComputedStyle; карточки как таблицы/картинки; выбор шрифта документа).
- 2026-10-25 PR #9 слит через PR #10 (общая сборка run 40226089437), стоит на VPS. См. [deeix-usage-line](deeix-usage-line.md).

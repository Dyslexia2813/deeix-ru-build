# deeix-word-revisions

_Правки (track changes), примечания, сноски, таблицы и номера пунктов Word (.docx и .doc) для модели — patches/word-revisions.patch, PR #23 (2026-11-08)._

- 2026-11-08 Kimi спросил, видит ли DEEIX правки и комментарии в договоре. Проверено тестом на v0.4.6 (backend/internal/infra/extraction/builtin/word.go, библиотека docxgo v2.4.0):
  - .docx: и вставленный (w:ins), и удалённый (w:del) текст выпадают целиком, примечания (comments.xml) не читаются. Модель видит только неизменённый текст. Версия Word 2007–365 не важна (формат один).
  - Ещё: таблицы .docx идут после всех абзацев, а не на своём месте.
  - .doc (наш чтец [deeix-doc-text](deeix-doc-text.md)): только основной текст (ccpText), примечания не берёт; удалённые исправления в .doc остаются в потоке текста → старое и новое склеиваются без пометок (подтверждено на .doc из LibreOffice: «3045»).
- Сделано 2026-11-08: patches/word-revisions.patch (нужен doc-text, ключ --no-word-revisions): word_docx_markup.go (.docx), word_doc_revisions.go + word_doc_lists.go (.doc: CHPX-ревизии, примечания, PAPX-таблицы, списки). Kimi выбрал «доучить чтец .doc» вместо LibreOffice в образе. Нужны автор и дата в пометках.
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/23 (слит 2026-11-08, 07a8a90; стоит на VPS, Kimi проверил — работает), сборка run 42490614167, инструкция ../instrukcii/word-revisions/ustanovka.md (маркер grep extractDocxMarkupText).
- Важно: тот же файл при повторной загрузке берётся из уже загруженных (дедуп по SHA, кроме failed) со старым текстом — сначала удалить в «Файлах».

# deeix-attachment-scroll

_Ряд вложений в поле ввода DEEIX — patches/attachment-scroll.patch (колесо вбок, полоса, счётчик файлов), PR #22._

- 2026-11-07 Kimi попросил: при 4+ вложениях дальние не видно. Выбрал: колесо мыши листает вбок, полоса прокрутки (не стрелки), автопрокрутка к новому файлу, счётчик = общее число файлов.
- Код: apps/web/features/chat/components/sections/chat-input.tsx (AttachmentGroup, у базового компонента scrollbar-none и snap-mandatory — полосу включать с `!`), хук features/chat/hooks/use-chat-attachment-row-scroll.ts (check:arch требует имя use-chat-*). Ключ chat.composer.attachmentTotal.
- Счётчик появляется только при переполнении; при его появлении ряд сужается, поэтому хук прокручивает к концу ещё раз.
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/22 (слит 2026-11-13, squash 9d5aa5e, стоит на VPS; апстрим v0.4.6). Ставится только заменой out. 2026-11-13: Kimi ставил #23–#26 без него (PR не был слит), ветка обновлена до main, сборка run 43482039949 (старая 42407320105), инструкция ../instrukcii/attachment-scroll/ustanovka.md (копия out-pered-vlozheniyami).
- Локальная проверка в браузере: собрать бэкенд go build ./cmd/server, CONFIG_FILE с frontend_dist_dir на apps/web/out, логин admin из лога, затем PATCH /api/v1/me/username и POST /api/v1/me/onboarding/complete {newPassword}, чтобы убрать онбординг. Playwright: ignoreDefaultArgs ['--hide-scrollbars'], иначе полос не видно.
- См. [deeix-table-fit](deeix-table-fit.md).

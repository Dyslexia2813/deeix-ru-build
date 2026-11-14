# deeix-composer-bottom

_Поле ввода внизу на пустом экране нового чата DEEIX (меню «/» и «@» раскрываются вверх) — patches/composer-bottom.patch, PR #13._

- 2026-10-26 Kimi: на пустом экране поле стояло посередине, меню «/» и «@» раскрывались вниз и были маленькими. Апстрим жёстко задаёт placementPreference "bottom" при !isConversationMode (chat-input.tsx, use-chat-mention-menu.ts, высота меню ≤340 px и ≤ места до края).
- Патч patches/composer-bottom.patch (флаг --no-composer-bottom, ставится последним): chat-empty.tsx — приветствие в flex-1 по центру, поле внизу (pb-3, как в диалоге); chat-input.tsx — все placementPreference = "top" (проп isConversationMode больше не читается).
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/13 (Kimi проверил на VPS, слит 2026-10-26, стоит на сервере), сборка run 40348888795 (зелёная), инструкция ../instrukcii/composer-bottom/ustanovka.md (копия out-pered-polem). Проверено Playwright 1280×800/600, 390×760.
- Локально: промпты для меню «/» создаются POST /api/v1/prompt-presets/mine {title,trigger,description,content}; модель для этого не нужна. Пароль superadmin — в логе сервера при первом запуске. См. [deeix-composer-triggers](deeix-composer-triggers.md).

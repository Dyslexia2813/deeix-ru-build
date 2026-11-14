# deeix-sidebar-width

_Ширина боковой панели DEEIX мышкой (перетаскивание правого края) — patches/sidebar-width.patch, PR #20._

- 2026-11-03 Kimi спросил, делали ли «вертикальную границу» подвижной — нет; 2026-11-06 сказал «да, давай сделаем».
- Патч patches/sidebar-width.patch (флаг --no-sidebar-width, ставится последним), только apps/web/components/ui/sidebar.tsx: SidebarWidthContext в SidebarProvider, --sidebar-width из state (по умолчанию 17.96875rem), localStorage `deeix.sidebar.width`; SidebarWidthHandle (w-2 у правого края sidebar-container, только при open и variant=sidebar), при перетаскивании пишет переменную прямо в sidebar-wrapper, data-width-dragging выключает transition; 220–560 px, чату ≥360 px; двойной щелчок сбрасывает. Подсказка через useLocale в коде, json переводов не трогаются (как в focus-toggle).
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/20 (Kimi проверил на VPS, слит 2026-11-07, squash b03c3bb, стоит на сервере), сборка run 42305192737, инструкция ../instrukcii/sidebar-width/ustanovka.md (копия out-pered-shirinoy).
- В этой сессии `pnpm install` апстрима запретил классификатор, локальная проверка в браузере не делалась — проверка типов только через сборку в Actions.

# deeix-sidebar-menu

_Компактное верхнее меню боковой панели DEEIX (всё кроме «Новый чат» под стрелкой ⌄) — patches/sidebar-menu.patch, PR #21._

- 2026-11-07 Kimi: пункты «Новый чат»…«Библиотека» занимают много места, мало остаётся под чаты/проекты; предложил наведение или «плюсик». Сделано по нажатию (наведение не работает на телефоне).
- Патч patches/sidebar-menu.patch (флаг --no-sidebar-menu, ставится последним, после sidebar-width): nav-main.tsx — «Новый чат» отдельно, справа кнопка ChevronDown (data-sidebar=nav-menu-toggle), остальное (Поиск + secondary) в grid-rows 0fr/1fr с inert; localStorage `deeix.sidebar.menu` ("open"/"closed"), по умолчанию свёрнуто; в режиме иконок (isCollapsed) всё видно. nav-main-item.tsx — проп `trailing` (рендер в SidebarMenuItem, Kbd mr-9). Подсказка через useLocale.
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/21 (Kimi проверил на VPS, слит 2026-11-07, squash c3d35e2, стоит на сервере), сборка run 42401926404, инструкция и скриншоты ../instrukcii/sidebar-menu/ (копия out-pered-menu).
- В этой сессии `pnpm install` и `next build` прошли локально; локальный запуск: deeix-chat с config.sqlite.example (пути заменить), онбординг Playwright-скриптом, проекты POST /api/v1/conversation-projects {name}, чаты POST /api/v1/conversations {title} (пустые чаты видны в «Недавних»). См. [deeix-sidebar-width](deeix-sidebar-width.md), [deeix-focus-toggle](deeix-focus-toggle.md).

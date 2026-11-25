# deeix-artifact-fullscreen

_HTML-страница (артефакт) из ответа модели открывается на весь экран одной кнопкой — patches/artifact-fullscreen.patch, PR #36 (сборка run 45702664478)._

- 2026-11-25 Kimi: если в ответе генерится HTML-страница (артефакт), её надо открывать на весь экран одной кнопкой — «как для таблицы ([deeix-md-table-full](deeix-md-table-full.md)), но не на всю ширину, а на весь экран».
- Как было: блок кода ```html (полная страница; короткие фрагменты рисуются на месте, см. [deeix-html-visual](deeix-html-visual.md)) показывается кодом, кнопка «Предпросмотр» (глаз) открывает узкую боковую панель «Артефакт» (`chat-artifact.tsx`).
- Сделано: над блоком кода рядом с глазом кнопка «На весь экран» (с подписью, `data-streamdown="code-block-fullscreen"`). Она передаёт `fullscreen: true` в `onOpenCodeArtifact`; `use-chat-artifacts.ts` хранит `fullscreenArtifact` отдельно от панели (панель не открывается). `ChatArtifactWorkspace` рисует один Dialog на всё окно (`data-artifact-fullscreen`): название из `<title>` страницы, иначе «HTML-страница»; «Скачать HTML»; крестик; Esc. Та же песочница iframe, что в панели.
- В шапке панели «Артефакт» тоже кнопка ⤢. Чтобы она не заезжала под вкладки в узкой панели, вкладки стоят в строке (не absolute по центру), а «Артефакт» и переключатель (если артефакт один) прячутся при ширине шапки < 34rem (container query).
- Подписи через `useLocale` (`shared/lib/artifact-fullscreen-labels.ts`), json не менялись. Патч от других не зависит (проверено на чистом v0.4.6), ставится последним. Ключ `--no-artifact-fullscreen`. Только сайт (`out`), строка для проверки: `code-block-fullscreen`.
- В `pnpm --filter @deeix/web check` старые ошибки не из этого патча (chat-cost-badge.tsx — biome; chat-pd-preview.tsx, use-chat-conversation-cost.ts — check:arch). typecheck чистый.

# deeix-app-icon

_Иконка сайта/PWA DEEIX (буква «D»): где лежат иконки, patches/app-icon.patch, PR #4, как обновить у установленного приложения._

- Иконки PWA: файлы apps/web/public/favicon.ico и public/pwa/*.png (фронтенд, out), адреса задаёт бэкенд (config.go defaults / config.yaml branding.*_url), манифест — бэкенд /api/v1/branding/manifest.webmanifest. Нужны и out, и бинарник.
- 2026-10-18: patches/app-icon.patch (флаг --no-app-icon) — бежевая «D» из картинки Kimi, maskable/apple на фоне #2a2520, к адресам добавлен ?v=d1 (при следующей смене иконок поднять до d2). PR https://github.com/Dyslexia2813/deeix-ru-build/pull/4, сборка run 39125176443 успешна; на VPS установлено, Kimi подтвердил; PR слит 2026-10-19.
- На VPS config.yaml (из шаблона DEEIX) явно задаёт branding.*_url без версии, поэтому ?v из config.go не действует; 2026-10-18 Kimi дана sed-команда дописать ?v=d1 в конфиг.
- iOS иконку не обновляет — только удалить и добавить заново; Chrome/Edge обновляют сами с задержкой. См. [deeix-deepseek-anthropic](deeix-deepseek-anthropic.md).

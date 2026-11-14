# tap-shell-app

_Android-оболочка DEEIX-RU (репо Dyslexia2813/tap-shell, WebView для tap.teamdream.online): устройство, сборка, кнопки._

- Репо Dyslexia2813/tap-shell: один MainActivity.java (WebView, только https tap.teamdream.online), layout activity_main.xml с плавающими круглыми кнопками («Обновить» и «Полный экран»).
- Сборка: GitHub Actions build.yml на push в main публикует APK в релиз `latest`; на ветке можно проверить через `gh workflow run build.yml --ref <ветка>` (релиз не публикуется). Локально Android SDK в облаке нет.
- 2026-10-23: кнопка «Обновить страницу» (web.reload) под кнопкой полного экрана — PR https://github.com/Dyslexia2813/tap-shell/pull/1 слит 2026-10-23, APK в релизе latest. Pull-to-refresh сознательно не делали (конфликт с прокруткой чата).
- 2026-10-23: обе кнопки перенесены в ряд вверху по центру (мешали элементам страницы в углу), версия 1.1; workflow берёт versionName и кладёт его в имя APK (deeix-ru-<версия>.apk) и заголовок релиза. PR https://github.com/Dyslexia2813/tap-shell/pull/2 слит 2026-10-23, релиз «DEEIX-RU 1.1». См. [apk-versioning](apk-versioning.md).
- 2026-10-23: полоса сверху в полном экране — окно не заходило под вырез камеры. Сделано edge-to-edge (setDecorFitsSystemWindows false, отступы и клавиатура вручную через insets), кнопки в полном экране ниже выреза. Версия 1.2, PR https://github.com/Dyslexia2813/tap-shell/pull/3 слит, релиз «DEEIX-RU 1.2»; полоса ушла (Kimi подтвердил).
- 2026-10-23: кнопки разведены по бокам от логотипа DEEIX в шапке сайта (marginTop 6dp, промежуток 84dp, по скриншоту 1080px), сдвиг ниже выреза убран. Версия 1.3, PR https://github.com/Dyslexia2813/tap-shell/pull/4 слит, релиз «DEEIX-RU 1.3»; следующая версия 1.4.

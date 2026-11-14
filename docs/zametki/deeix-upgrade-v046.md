# deeix-upgrade-v046

_Переход сборки DEEIX-RU на v0.4.6 (2026-11-03): все 12 патчей легли без правок, PR #18, инструкция для VPS._

- 2026-11-03 Kimi попросил поставить «26-ю версию» = апстрим v0.4.6 (тег от 2026-10-22, последний на тот день).
- Все 12 патчей install.sh ложатся на v0.4.6 без изменений (по отдельности `--check` ругается на turn-resources и focus-toggle — они зависят от предыдущих патчей, это норм; сам install проходит). Новых строк для перевода 0. go test (108 ok) и `pnpm --filter @deeix/web check` чистые. В v0.4.6 go.mod требует Go 1.26 (GOTOOLCHAIN=auto).
- PR https://github.com/Dyslexia2813/deeix-ru-build/pull/18 слит 2026-11-05 (Kimi: «сливаем»), сборка run 41639683403. Инструкция: ../instrukcii/deeix-update/obnovlenie-v0.4.6.md (копия базы в /opt/deeix-ru/baza-pered-0.4.6.tar.gz, старая версия в *-old).
- См. [deeix-upgrade-v045](deeix-upgrade-v045.md), [deeix-db-backup](deeix-db-backup.md).

# deeix-db-backup

_Резервная копия базы DEEIX на VPS — команда tar тома deeix-chat-app-data в /opt/deeix-backup, инструкция и восстановление._

- 2026-10-22 старая копия data-0.4.4 удалена при уборке /opt/deeix-ru; Kimi попросил сделать копию БД. Других бэкапов не было.
- Способ: остановить app, `tar -czf /opt/deeix-backup/deeix-data-ДАТА.tar.gz -C /var/lib/docker/volumes/deeix-chat-app-data/_data .`, запустить app. Инструкция с восстановлением: ../instrukcii/db-backup/kopiya-bd.md.
- 2026-10-22 Kimi согласился на автокопию: дана команда — скрипт /opt/deeix-backup/backup.sh, /etc/cron.d/deeix-backup (04:00 по времени сервера), хранить 7, журнал backup.log. Пробный запуск 2026-10-22 20:36 UTC успешен (14 МБ); сервер в UTC (04:00 UTC = 07:00 МСК); cron включён (был inactive); первая ночная копия 2026-10-23 04:00 UTC успешна. Снимок 1Panel: файлов *snapshot*.tar.gz на диске не найдено; deeix-chat-app запущен вручную (не приложение 1Panel), его том в снимок скорее всего не входит.
- См. [deeix-table-fit](deeix-table-fit.md), [deeix-upgrade-v045](deeix-upgrade-v045.md).

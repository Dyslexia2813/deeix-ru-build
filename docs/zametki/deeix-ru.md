# deeix-ru

_Проект DEEIX-RU: доработка DEEIX-Chat на VPS — русификация (deeix-ru-build), Android-оболочка tap-shell, скиллы ru-legal и MCP pravo-mcp._

- [stated] проект: доработка и настройка https://github.com/DEEIX-AI на собственном VPS; сайт с DEEIX-Chat: tap.teamdream.online
- [stated] этап 1: русификация и мелкие правки — https://github.com/Dyslexia2813/deeix-ru-build
- [stated] этап 2: Android-оболочка для сайта — https://github.com/Dyslexia2813/tap-shell
- [stated] этап 3 (текущий): импортировать в DEEIX-Chat скиллы по договорному праву (папка contract-law) из https://github.com/AlsKozlov/ru-legal/ и подключить MCP-сервер pravo-mcp (обязателен для этих скиллов; также нужен в DEEIX-Chat как самостоятельный инструмент)
- [stated] в проекте 4 чата; хочет объединить память между ними, рассматривает: ссылки на репозитории в Инструкциях или загрузка репозиториев/их частей в Файлы
- [stated] VPS: Ubuntu 24.04 amd64, ядро 6.8.0-136-generic
- [stated] состояние проекта ведётся в едином файле STATE.md в Файлах проекта (сводка из трёх чатов от 2026-10-13)
- [stated] 2026-10-14: pravo-mcp поднят на VPS и подключён к DEEIX-Chat, поиск через инструмент работает; следующий шаг — пробный импорт одного скилла contract-law
- [stated] Cloudflare с tap.teamdream.online не используется
- [stated] 2026-10-14: разбирает ошибку «Генерация прервана / Файл ещё не готов» в DEEIX-Chat при отправке двух PDF (договор 5,3 МБ без текстового слоя и ДС); повторяется при повторной отправке

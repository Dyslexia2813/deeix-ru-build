# Установка: HTML-страница из ответа на весь экран (PR #36)

PR: https://github.com/Dyslexia2813/deeix-ru-build/pull/36
Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/45702664478 (ветка `claude/project-thread-yvvknj`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются. Сборка содержит всё, что уже стоит на сервере (main), плюс эту правку.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-36`, ставит новый, проверяет, что в нём новый код, и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-36 && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-36 && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && echo "новый код: $(grep -rl code-block-fullscreen /opt/deeix-ru/out/_next | wc -l)" && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

Должно быть «новый код: 1» (или больше) и в конце `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте чат, где модель написала HTML-страницу (блок кода `html`), или попросите новую: «Сделай одностраничный сайт для юридического бюро».

- Над блоком кода справа, рядом с глазом «Предпросмотр», появилась кнопка «⤢ На весь экран».
- Нажмите её: страница откроется на весь экран. Сверху её название, кнопка «Скачать HTML» и крестик. Закрыть — крестиком или Esc.
- Если открыть страницу в боковой панели (глаз), в её шапке тоже есть ⤢ — открывает ту же страницу на весь экран.

Примеры: `pk-knopka.png`, `pk-polnyj-ekran.png`, `pk-panel.png`, `telefon-knopka.png`, `telefon-polnyj-ekran.png` в этой папке.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-36 /opt/deeix-ru/out && $C up -d --force-recreate app
```

## Уборка (когда всё хорошо)
```
clear && rm -rf /opt/deeix-ru/out-pered-36 && echo готово
```

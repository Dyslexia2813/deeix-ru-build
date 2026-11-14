# Установка строки «Скиллы / Инструменты» под ответом (PR #10, вместе с PR #9)

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/40226089437 (ветка `feat/turn-resources`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются.

В сборке все слитые патчи (таблицы, HTML-вёрстка, кнопки «/» и «@») и новый экспорт в Word со шрифтами с экрана (PR #9). Если PR #9 уже стоит, эту сборку можно просто поставить поверх.

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (старый `deeix-ru-out.zip` замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-stroka`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-stroka && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-stroka && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Откройте ответ, где были скиллы или вызовы инструментов (подойдут и старые). Под строкой с моделью и ценой появится:

- «Скиллы: …» — скиллы, подключённые к запросу (наведите курсор, чтобы увидеть триггеры);
- «Инструменты: search_npa ×3, …» — что модель реально вызвала и сколько раз (в подсказке видно, сколько вызовов закончились ошибкой).

Если в запросе не было ни скиллов, ни вызовов, строки нет. Экспорт в Word должен сохранять шрифты и цвета, как после PR #9.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-stroka /opt/deeix-ru/out && $C up -d --force-recreate app
```

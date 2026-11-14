# Установка: «/» и «@» в меню кнопки «+» (PR #12) вместе с PR #11

Сборка: https://github.com/Dyslexia2813/deeix-ru-build/actions/runs/40317612867 (ветка `build/plus-and-glm`).
Меняется только сайт (`out`), бэкенд и база не затрагиваются.

Сборка содержит всё, что уже слито в main (HTML-вёрстка, широкие таблицы, экспорт в Word со шрифтами, скиллы и инструменты под ответом), новое меню «+» (PR #12) и подсветку кнопки рассуждения у GLM (PR #11).

## Шаг 1. Скачать
1. Откройте ссылку на сборку, внизу в разделе Artifacts скачайте только `deeix-ru-out` (придёт `deeix-ru-out.zip`).
2. В 1Panel → «Файлы» загрузите его в `/opt/deeix-ru/` (если там лежит старый `deeix-ru-out.zip`, замените).

## Шаг 2. Одна команда в терминале 1Panel
Сохраняет текущий сайт в `out-pered-knopkami`, ставит новый и перезапускает контейнер.

```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out-pered-knopkami && mv /opt/deeix-ru/out /opt/deeix-ru/out-pered-knopkami && mkdir /opt/deeix-ru/out && unzip -q /opt/deeix-ru/deeix-ru-out.zip -d /opt/deeix-ru/out && $C up -d --force-recreate app && sleep 10 && curl -s http://127.0.0.1:8080/healthz
```

В конце должно быть `{"status":"ok",...}`.

## Шаг 3. Проверить
Обновите страницу (Ctrl+F5, в приложении на телефоне кнопка «Обновить»). Кнопок «/» и «@» в нижнем ряду больше нет. Нажмите «+»: под «Загрузить файл» и «Снимок экрана» будут «/ Скиллы и промпты» и «@ Файлы, модели, инструменты». Выберите первый пункт: в поле встанет «/», откроется меню скиллов, а на телефоне клавиатура. У модели GLM значок рассуждения должен гореть на любом уровне, в том числе по умолчанию.

## Откат
```
clear && cd /opt/deeix-chat/DEEIX-Chat/deploy && C="docker compose -f docker-compose.sqlite.yml -f /opt/deeix-ru/compose.override.yml" && rm -rf /opt/deeix-ru/out && mv /opt/deeix-ru/out-pered-knopkami /opt/deeix-ru/out && $C up -d --force-recreate app
```

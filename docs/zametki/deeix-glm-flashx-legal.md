# deeix-glm-flashx-legal

_Параметры модели zai/glm-5.3-flashx-legal в DEEIX (иконка «chatglm»): JSON возможностей, рассуждение не объявлять._

- 2026-11-11 Kimi добавил модель `zai/glm-5.3-flashx-legal` (в DEEIX иконка/вендор «chatglm») и попросил заполнить «Возможности модели».
- По сторонним каталогам (официальную страницу Z.ai не открыть из облака): вход текст+картинки+видео, контекст ~1M, мышление обязательное (выключить нельзя), function calling, цена $0,37 вход / $0,075 кэш / $1,25 выход за 1M.
- Схема DEEIX v0.4.6 (backend/internal/domain/channel): `inputModalities` (text,image,pdf,audio,video), `contextWindow`, `maxOutputTokens`, `reasoning` {format, levels, default}. Формат `toggle` шлёт `thinking.type` enabled/disabled и требует ровно levels ["none","high"]; `openai` шлёт `reasoning_effort`.
- Рекомендовано: `{"inputModalities":["text","image"],"contextWindow":1000000}`, блок `reasoning` не объявлять (мышление и так всегда включено; «Выкл.» могло бы дать ошибку). Запись кэша в цене = цене входа ([deeix-model-catalog](deeix-model-catalog.md)). Веб-поиск — через MCP ([deeix-glm-web-search](deeix-glm-web-search.md), [deeix-deepseek-search](deeix-deepseek-search.md)). С картинками у модели своё OCR не нужно ([deeix-own-ocr](deeix-own-ocr.md)).
- 2026-11-13 Kimi: вопрос решён, тред закрыт.

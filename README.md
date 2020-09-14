# Описание

API предназначено для получения расписания групп АГУ

## Требования

Для работы нужен интерпретатор Python версии 3+, docker, docker-compose, make

## Getting Started
```bash
git clone https://github.com/Wishpering/asu_rasp_api.git
```
Нужно заполнить конфигурационный файл в папке ./configs и поменять ENV для API в ./docker/.env


## Возможности

Для авторизации используются Headers, а именно поле Authorization.

Поддерживаются простые GET-запросы, а именно:

Запрос используется для получения токена для дальнейшей работы с API:
```http
- http://server/token
  - Headers: 
    - Authorization: {{ password = пароль из ./docker/.env }}
```

Запрос используется для проверки валидности токена:
```http
- http://server/token/check
  - Headers: 
    - Authorization: {{ token_for_check }}
```

Запрос для получения расписания группы:
```http
- http://server/rasp
  - Headers:
    - Authorization: {{ token }}
  - Параметры:
    - id = номер группы
    - date = дата, для которой будет проверяться расписание. Формат - yy:mm:dd.
    - end_date = дата, до которой будет проверяться расписание. Формат - yy:mm:dd.
```

Запрос для получения списка групп:
```http
- http://server/pool
  - Headers: 
    - Authorization: {{ token }}
```
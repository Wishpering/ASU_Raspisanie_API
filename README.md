# Описание

API предназначено для получения расписания групп/преподавателей АГУ

## Требования

Для работы нужен интерпретатор Python версии 3+, docker, docker-Compose, make, go

## Getting Started
```bash
git clone https://github.com/Wishpering/asu_rasp_api.git
```
Нужно заполнить конфигурационный файл в папке /configs.

Так же необходимо указать группы и фамилии преподавателей для добавления в пулл в формате: \
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;- Группы - создать файл data/groups и заполнить его в формате:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; ```"факультет:номер_группы"``` \
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;- Преподаватели - создать файл data/preps и заполнить его в формате: \
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;```"факультет:кафедра:фамилия"```

## Возможности

Для авторизации используются Headers, а именно поле Authorization.

Поддерживаются простые GET-запросы, а именно:

Запрос используется для получения токена для дальнейшей работы с API:
```http
- http://server/token
  - Headers: 
    - Authorization:password = пароль из файла config.json, предназначенный для генерации токена
```

Запрос для получения расписания группы:
```http
- http://server/rasp/groups
  - Headers:
    - Authorization:token
  - Параметры:
    - id = номер группы
    - date = дата, для которой будет проверяться расписание. Формат - yy:mm:dd.
    - end_date = дата, до которой будет проверяться расписание. Формат - yy:mm:dd.
```

Запрос для получения расписания преподавателя:
```http
- http://server/rasp/preps
  - Headers:
    - Authorization:token
  - Параметры:
    - name = фамилия/имя преподавателя
    - date = дата, для которой будет проверяться расписание. Формат - yy:mm:dd.
    - end_date = дата, до которой будет проверяться расписание. Формат - yy:mm:dd.
```

Запрос для получения списка групп:
```http
- http://server/pool/groups
  - Headers: 
    - Authorization:token
```

Запрос для получения списка преподавателей:
```http
- http://server/pool/preps
  - Headers:
    - Authorization:token
```
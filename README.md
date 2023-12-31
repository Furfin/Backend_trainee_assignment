# Тестовое задание для стажёра Backend
# Сервис динамического сегментирования пользователей


## Запуск
Проект разрабатывался на Ubuntu 22.04.03 LTS

Для запуска приложения:

> `make start`

Для запуска тестов:

> `make test`


## Описание эндпоинтов

- [POST] "/segment"
  - Метод создания нового сегмента
  - На вход принимает json
  - `{"slug": "string","upadd": 0}`
  - slug - обязательный параметр, является уникальным для каждго сегмента
  - upadd - необязательный параметр, описывает процент пользователей автоматически добавляемых в сегмент при его создании, должен быть целым числом больше и равным 0 
  - Пример вывода
    ```json
    {
        "status": "Ok",
        "error": ""
    }
    ```
- [DELETE] "/segment"
  - Метод удаления сегмента
  - На вход принимает json
  - `{"slug": "string"}`
  - При удалении сегмента, удаляются так же и записи о вхождении пользователей в этот сегмент
  - Пример вывода
    ```json
    {
        "status": "Ok",
        "error": ""
    }
    ```
- [GET] "/user/{userid}"
  - Метод получения информации о пользователе
  - На вход число - id пользователя
  - При отсуствии пользователя в БД создается запись о новом пользователе
  - Пример вывода
    ```json 
    { "status": "Ok",
    "error": "User info",
    "Segments": [{
            "ID": 3,
            "CreatedAt": "2023-08-30T10:59:01.604195Z",
            "UpdatedAt": "2023-08-30T10:59:01.604195Z",
            "DeletedAt": null,
            "Slug": "strin"
        }
    ]}
    ```
- [POST] "/user/{userid}/add"
  - Метод добавления и удаления пользователя из сегментов
  - На вход число - id пользователя
  - json
    ```json
        {
        "AddTo": [
                    "seg1"
        ],
        "RemoveFrom": [
                    "seg2"
        ],
        "ttl_days": {
            "seg1":10,
            "seg2":5,
            "seg3":15
        }
    }
    ```
  - При отсуствии пользователя в БД создается запись о новом пользователе
  - Пример вывода
    ```json 
        {
            "status": "Ok",
            "error": ""
        }
    ```
  - Все поля запроса можно оставить пустыми
  - После запроса, указанный пользователь добавляется в указанные сегменты, а после удаляется из также указанных сегментов
  - Записи об удалении и добавлении пользователей в сегмент не стираются для генерации истории пользователя
- [POST] "/user/{userid}/csv"
  - Метод получения информации о пользователе
  - На вход число - id пользователя
  - json
    ```json
      {
          "month": 8,
          "year": 2023
      }
    ```
  - При отсуствии пользователя в БД создается запись о новом пользователе
  - Ответом отсылается файл history.csv с нужными данными за указанный пероид
  - Все поля запроса обязательны для заполнения и описывают левую границу поиска
## Вопросы по ТЗ
- Не указан механизм добавления пользователей
  - Решение:
  >> В каждом запросе, в котором фигурируют пользователи, я (при отсутствии оного в ДБ) добавляю нового пользователя, userid - всегда уникальный

## Интересные факты
- TTL для записей вхождения пользователей в сегмент задается в днях. При запуске сервера запускается горутина, которая каждые 15 минут проверяет записи в БД и удаляет те, существование которых превышает данное им время
- (Со стандартными настройками) Swagger запускается по адресу 
  >> http://localhost:8084/swagger/index.html
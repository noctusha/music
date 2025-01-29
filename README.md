# Music Library API 🎵

REST API для управления онлайн-библиотекой песен с интеграцией внешнего источника данных

## Особенности
- Полноценное CRUD-управление песнями
- Интеграция с внешним API для автоматического дополнения данных
- Пагинация и фильтрация результатов
- Автогенерация Swagger-документации
- Автомиграции для PostgreSQL
- Конфигурация через .env файл
- Подробное логирование

## Быстрый старт

### Требования
- Go 1.21+
- PostgreSQL 15+
- Docker (опционально)

### Установка
1. Клонировать репозиторий:
```
git clone https://github.com/noctusha/music.git
cd music
```

2. Установить зависимости:
```
go mod download
```

3. Создать файл конфигурации:
```
cp .env.example .env
```

4. Запустить сервис с миграциями:
```
go run main.go migrate up
go run main.go serve
```

## Конфигурация (.env)
```
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=music_library
MUSIC_INFO_API=http://api.music-info.com/v1
LOG_LEVEL=debug
```

## API Endpoints (основные методы)

1. `GET /api/songs` - получение списка песен с фильтрацией
   Параметры: group, song, year, page, limit. 

   Пример: ``GET /api/songs?group=Muse&page=1&limit=10``

2. `GET /api/songs/{id}/text` - получение текста песни с пагинацией
   Параметры: page, limit.

   Пример: ``GET /api/songs/123/text?page=2&limit=5``

3. `POST /api/songs` - добавление новой песни
   ```
   {
    "group": "Muse",
    "song": "Supermassive Black Hole"
   }
   ```

4. `PUT /api/songs/{id}` - обновление данных песни
   

5. `DELETE /api/songs/{id}` - удаление песни


## Структура БД

Таблицы создаются автоматически через миграции:
```
CREATE TABLE songs (
    id SERIAL PRIMARY KEY,
    group VARCHAR(255) NOT NULL,
    song VARCHAR(255) NOT NULL,
    release_date DATE,
    text TEXT,
    youtube_link VARCHAR(512),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

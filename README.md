# Сервис игры в кости

Простой игровой сервис на Go с gRPC API, где пользователь вызывает метод для игры в «подбрасывание кубика». Сервис генерирует два случайных результата (от 1 до 6): один для игрока, второй — для сервера. Победитель определяется по наибольшему числу, а результат игры записывается в базу данных PostgreSQL.

## Особенности

- Реализация Clean Architecture с полным разделением бизнес-логики от инфраструктуры
- gRPC API для игры и проверки результатов
- Несколько генераторов случайных чисел с удобным интерфейсом
- Реализация алгоритма Provably Fair для проверки честности результатов
- PostgreSQL база данных для хранения результатов игр
- Поддержка Docker и docker-compose для простого развертывания

## Архитектура

Проект следует принципам Clean Architecture:

- **Domain**: Содержит бизнес-логику, модели и интерфейсы
- **Use Cases**: Прикладные бизнес-правила
- **Infrastructure**: Внешние фреймворки, базы данных и сервисы

## Технический стек

- Go 1.24+
- gRPC
- PostgreSQL
- Docker & Docker Compose

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)
- PostgreSQL (для локальной разработки без Docker)

### Использование Docker Compose (рекомендуется)

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/NikolayAleshin/dice-game.git
   cd dice-game
   ```

2. Соберите и запустите сервис:
   ```bash
   docker-compose up -d
   ```

3. Сервис будет доступен по адресу `localhost:9090`

### Локальная настройка для разработки

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/NikolayAleshin/dice-game.git
   cd dice-game
   ```

2. Установите зависимости:
   ```bash
   go mod download
   ```
   
3. Выполните генерацию protobuf и gRPC кода выполнив скрипт
   ```bash
    ./scripts/generate_proto.sh      
   ```

4. Запустите PostgreSQL (можно использовать Docker):
   ```bash
    docker run -d --name dice-game-postgres \                                                                  
   -p 5432:5432 \
   -e POSTGRES_USER=postgresql \
   -e POSTGRES_PASSWORD=password123 \
   -e POSTGRES_DB=dice_game_db \
   postgres:14-alpine

   ```

5. Настройте приложение (host localhost использовать, создайте файл config.yaml или используйте переменные окружения):
   ```yaml
   host: localhost
   ```

6. Запустите приложение:
   ```bash
   go run cmd/main.go
   ```

## Использование

### Игра в кости

Вы можете использовать `grpcurl` для тестирования сервиса:

```bash
grpcurl -plaintext -d '{"player_id": "player123"}' localhost:9090 dice_game.DiceGameService/Play
```

Пример ответа:
```json
{
  "gameId": "d7d2c2b2-36a7-4566-adda-1e5f4250d398",
  "playerDice": 4,
  "serverDice": 2,
  "winner": "PLAYER",
  "playedAt": "2025-03-16T01:26:25+04:00",
  "generatorUsed": "provably_fair",
  "verificationKey": "server-seed-1742073977829982000:2:cc483ba185f086eddf64d3bbbeddf3521259395278b02ee7992d0b50edbca182"
}
```

### Проверка результата игры

Для проверки результата игры (для игр с Provably Fair):

```bash
grpcurl -plaintext -d '{"game_id": "d7d2c2b2-36a7-4566-adda-1e5f4250d398", "client_seed": "2025-03-16T01:26:25+04:00"}' localhost:9090 dice_game.DiceGameService/Verify
```

Примечание: В текущей реализации клиентский seed - это временная метка из поля `playedAt` в ответе Play.

## Как работает Provably Fair

1. Сервер генерирует серверный seed
2. Когда происходит игра, система:
   - Берет текущую метку времени как клиентский seed
   - Комбинирует серверный seed + клиентский seed + nonce
   - Генерирует SHA-256 хеш
   - Выводит случайные числа из хеша
3. Для проверки:
   - Предоставьте ID игры и клиентский seed (метку времени)
   - Система воссоздаст хеш и сравнит полученные числа

## Добавление новых генераторов случайных чисел

Чтобы добавить новый генератор случайных чисел:

1. Реализуйте интерфейс Generator:
```go
type Generator interface {
    // Generate генерирует случайное число между min и max
    Generate(min, max int) (int, error)
    
    // Name возвращает имя генератора
    Name() string
}
```

2. Добавьте ваш новый генератор в список в `pkg/app/app.go`:
```go
func (a *Application) initRandomGenerators() {
    randomGenerators := []random.Generator{
        random.NewStandardGenerator(),
        random.NewCryptoGenerator(),
        random.NewYourCustomGenerator(), // Добавьте ваш генератор здесь
    }
    
    a.randomService = service.NewRandomService(randomGenerators)
}
```
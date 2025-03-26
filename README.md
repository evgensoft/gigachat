# GigaChat API Client

Go-клиент для работы с API GigaChat.

## Установка

```bash
go get github.com/yourusername/gigachat
```

## Использование

### Получение токена

```go
package main

import (
    "fmt"
    "log"
    "gigachat"
)

func main() {
    // Создаем клиент (по умолчанию используется ScopePersonal)
    client := gigachat.NewClient("your-client-id", "your-client-secret")

    // При необходимости можно изменить scope
    client.SetScope(gigachat.ScopeB2B)

    // Получаем токен
    token, err := client.GetToken()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Токен: %s\n", token.AccessToken)
    fmt.Printf("Срок действия до: %d\n", token.ExpiresAt)
}
```

### Использование чата

```go
package main

import (
    "fmt"
    "log"
    "gigachat"
)

func main() {
    // Создаем клиент
    client := gigachat.NewClient("your-client-id", "your-client-secret")

    // Создаем запрос к чату
    req := &gigachat.ChatRequest{
        Model: gigachat.ModelGigaChat,
        Messages: []gigachat.Message{
            {
                Role:    gigachat.RoleSystem,
                Content: "Ты - полезный ассистент",
            },
            {
                Role:    gigachat.RoleUser,
                Content: "Привет! Как дела?",
            },
        },
    }

    // Отправляем запрос
    resp, err := client.Chat(req)
    if err != nil {
        log.Fatal(err)
    }

    // Выводим ответ
    for _, choice := range resp.Choices {
        fmt.Printf("Ответ: %s\n", choice.Message.Content)
        fmt.Printf("Использовано токенов: %d\n", resp.Usage.TotalTokens)
    }
}
```

## Константы для scope

- `ScopePersonal` - для физических лиц (используется по умолчанию)
- `ScopeB2B` - для ИП и юридических лиц (платные пакеты)
- `ScopeCorp` - для ИП и юридических лиц (pay-as-you-go)

## Доступные модели

- `ModelGigaChat` - базовая модель
- `ModelGigaChatPro` - профессиональная модель
- `ModelGigaChatMax` - максимальная модель

## Роли сообщений

- `RoleSystem` - системный промпт
- `RoleUser` - сообщение пользователя
- `RoleAssistant` - ответ ассистента
- `RoleFunction` - результат работы функции

## Обработка ошибок

Клиент возвращает структурированные ошибки типа `APIError`:

```go
if err != nil {
    if apiErr, ok := err.(*gigachat.APIError); ok {
        fmt.Printf("Ошибка API: %d - %s\n", apiErr.Code, apiErr.Message)
    } else {
        fmt.Printf("Другая ошибка: %v\n", err)
    }
}
```

## Требования

- Go 1.22 или выше 
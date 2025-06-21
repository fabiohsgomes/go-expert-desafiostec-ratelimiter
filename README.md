# RateLimiter em Go

Um Ratelimiter flexível e configurável para aplicações Go que suporta limitação baseada em IP e em token, com armazenamento em Redis.

## Funcionalidades

- Limitação de taxa baseada em IP
- Limitação de taxa baseada em token com limites configuráveis por token
- Armazenamento em Redis com suporte a outros backends de armazenamento através de interface
- Configurável através de variáveis de ambiente ou arquivo .env
- Suporte a middleware HTTP
- Duração de bloqueio configurável para limites excedidos
- Limites de token têm precedência sobre limites de IP quando ambos estão presentes

## Instalação

```bash
go get github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter
```

## Configuração

O ratelimiter pode ser configurado através de variáveis de ambiente ou um arquivo .env:

```env
# Configurações gerais de limitação de taxa
RATE_LIMIT_MAX_REQUESTS=10        # Requisições por segundo padrão
RATE_LIMIT_BLOCK_DURATION=5m      # Duração do bloqueio após limite excedido
RATE_LIMIT_TOKEN_HEADER=API_KEY   # Nome do cabeçalho para tokens de API

# Configuração do Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Uso

### Uso Básico com Servidor HTTP

```go
package main

import (
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/internal/config"
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/internal/middleware"
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

func main() {
    // Carrega configuração
    cfg, _ := config.LoadConfig()

    // Inicializa armazenamento Redis
    store, _ := storage.NewRedisStorage("localhost:6379", "", 0)
    defer store.Close()

    // Cria ratelimiter
    limiter := ratelimiter.New(store, cfg)

    // Cria middleware
    rateLimiterMiddleware := middleware.New(limiter, cfg)

    // Hander http
    http.Handle("/", rateLimiterMiddleware.Handler(yourHandler))
    http.ListenAndServe(":8080", nil)
}
```

### Configurando Limites de Token

```go
cfg := ratelimiter.NewConfig()
cfg.SetTokenLimit("abc123", 100, time.Minute*5)  // 100 requisições por segundo, bloqueio de 5 minutos
```

## Configuração do Redis

Um arquivo docker-compose.yml é fornecido para facilitar a configuração do Redis:

```bash
docker-compose up -d
```

## Códigos de Resposta

- 200: OK
- 429: Too Many Requests
- 500: Internal Server Error

Quando o limite de requisições é excedido, a resposta incluirá:
```json
{
    "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

## Detalhes de Implementação

O ratelimiter usa Redis para rastrear contagens de requisições e status de bloqueio:
- Contagens de requisições são armazenadas com expiração de 1 segundo
- Status de bloqueio é armazenado pela duração de bloqueio configurada
- Limites de token têm precedência sobre limites de IP quando ambos estão presentes
- O sistema é projetado para ser thread-safe e distribuído

## Testes

Para executar os testes:

```bash
go test ./...
```
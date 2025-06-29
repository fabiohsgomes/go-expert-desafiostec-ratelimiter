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

O ratelimiter pode ser configurado através de variáveis de ambiente ou um arquivo .env. A configuração é carregada automaticamente através das funções `middleware.LoadConfig()` e `middleware.LoadRedisConfig()`:

```env
# Configurações gerais de limitação de taxa
RATE_LIMIT_MAX_REQUESTS=10        # Requisições por segundo padrão
RATE_LIMIT_BLOCK_DURATION=5m      # Duração do bloqueio após limite excedido
RATE_LIMIT_TOKEN_HEADER=API_KEY   # Nome do cabeçalho para tokens de API

# Configuração do Redis
REDIS_ADDR=localhost:6379         # Endereço do servidor Redis
REDIS_PASSWORD=                   # Senha do Redis (opcional)
REDIS_DB=0                       # Número do banco de dados Redis

# Configurações avançadas
RATE_LIMIT_ENABLED=true          # Habilita/desabilita o rate limiting
RATE_LIMIT_CLEANUP_INTERVAL=5m   # Intervalo de limpeza de registros expirados
```

As configurações podem ser carregadas programaticamente:

```go
// Carrega configuração do rate limiter
cfg, err := middleware.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Carrega configuração do Redis
addr, password, db := middleware.LoadRedisConfig()
```

## Uso

### Uso Básico com Servidor HTTP

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/middleware"
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/ratelimiter"
    "github.com/fabiohsgomes/go-expert-desafiostec-ratelimiter/pkg/storage"
)

func main() {
    // Carrega configuração
    cfg, err := middleware.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Carrega limites de token das variáveis de ambiente
    cfg.LoadTokenLimitsFromEnv()

    // Inicializa armazenamento Redis
    addr, password, db := middleware.LoadRedisConfig()
    store, err := storage.NewRedisStorage(addr, password, db)
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    // Cria ratelimiter
    limiter := ratelimiter.New(store, cfg)

    // Cria middleware
    rateLimiterMiddleware := middleware.New(limiter, cfg)

    // Cria um handler simples
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    // Aplica o middleware ao handler
    http.Handle("/", rateLimiterMiddleware.Handler(handler))

    // Inicia o servidor
    log.Println("Servidor iniciando na porta :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
```

### Configurando Limites de Token

Os limites de token podem ser configurados através de variáveis de ambiente:

```env
# Formato: TOKEN_LIMIT_<TOKEN>=<requests>:<duration>
TOKEN_LIMIT_ABC123=100:5m         # 100 req/s, bloqueio 5min
TOKEN_LIMIT_XYZ789=50:10m         # 50 req/s, bloqueio 10min
TOKEN_LIMIT_PREMIUM=1000:1m       # 1000 req/s, bloqueio 1min
```

Ou programaticamente:

```go
cfg := ratelimiter.NewConfig()
cfg.LoadTokenLimitsFromEnv()  // Carrega das variáveis de ambiente
// ou
cfg.SetTokenLimit("abc123", 100, time.Minute*5)  // Configuração manual
```

## Executando com Docker

Um arquivo docker-compose.yml é fornecido para executar a aplicação completa:

```bash
# Executar aplicação e Redis
docker-compose up -d

# Apenas Redis para desenvolvimento
docker-compose up -d redis
```

A aplicação estará disponível em http://localhost:8080

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

Para informações detalhadas sobre testes de unidade e integração, consulte o arquivo [TESTING.md](TESTING.md).

```bash
# Executar todos os testes
go test ./... -v

# Executar testes de integração com Redis
go test ./test -tags=integration -v
```
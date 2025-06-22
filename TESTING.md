# Testes de Integração - Rate Limiter

Este documento descreve os testes de integração implementados para o middleware de rate limiter.

## Tipos de Testes

### 1. Testes de Middleware (`internal/middleware/integration_test.go`)

Testa o middleware isoladamente com storage em memória:

```bash
# Executar todos os testes do middleware
go test ./internal/middleware -v

# Executar apenas testes de integração
go test ./internal/middleware -v -run TestRateLimiterMiddleware_Integration

# Executar testes de tratamento de erros
go test ./internal/middleware -v -run TestRateLimiterMiddleware_ErrorHandling
```

**Cenários testados:**
- Rate limiting baseado em IP
- Rate limiting baseado em token
- Precedência de token sobre IP
- IPs diferentes têm limites separados
- Reset após duração do bloqueio
- Respeito aos headers X-Forwarded-For e X-Real-IP
- Tratamento de erros internos

### 2. Testes de Integração Completa (`test/integration_test.go`)

Simula um servidor HTTP real com diferentes cenários:

```bash
# Executar teste de integração completa
go test ./test -v -run TestFullIntegration

# Executar benchmark de performance
go test ./test -bench=BenchmarkRateLimiterMiddleware -benchmem
```

**Cenários testados:**
- Servidor HTTP completo com middleware
- Diferentes tipos de tokens (premium, basic)
- Limites específicos por token
- Reset de limites após bloqueio
- Performance sob carga

### 3. Testes com Redis (`test/redis_integration_test.go`)

Testa integração com Redis real (requer Redis rodando):

```bash
# Executar testes com Redis (requer tag integration)
go test ./test -tags=integration -v -run TestRedisIntegration

# Com Redis customizado
REDIS_ADDR=localhost:6379 go test ./test -tags=integration -v -run TestRedisIntegration

# Benchmark com Redis
go test ./test -tags=integration -bench=BenchmarkRedisRateLimiter -benchmem
```

**Cenários testados:**
- Rate limiting com persistência no Redis
- Persistência entre instâncias do middleware
- Cleanup após duração do bloqueio
- Performance com Redis

## Configuração do Ambiente

### Para testes básicos (sem Redis)
Não requer configuração adicional. Os testes usam storage em memória.

### Para testes com Redis
1. Instalar e iniciar Redis:
```bash
# Docker
docker run -d -p 6379:6379 redis:alpine

# Ou usando docker-compose (se disponível no projeto)
docker-compose up -d redis
```

2. Definir variável de ambiente (opcional):
```bash
export REDIS_ADDR=localhost:6379
```

## Executar Todos os Testes

```bash
# Testes básicos (sem Redis)
go test ./... -v

# Incluindo testes com Redis
go test ./... -v
go test ./test -tags=integration -v

# Com coverage
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Resultados Esperados

### Performance Benchmarks
- **Memory Storage**: ~3000 ns/op, ~5656 B/op, 18 allocs/op
- **Redis Storage**: Varia conforme latência da rede

### Cobertura de Testes
Os testes cobrem:
- ✅ Rate limiting por IP
- ✅ Rate limiting por token
- ✅ Precedência de token
- ✅ Headers de proxy (X-Forwarded-For, X-Real-IP)
- ✅ Tratamento de erros
- ✅ Reset após bloqueio
- ✅ Persistência (Redis)
- ✅ Performance

## Estrutura dos Testes

```
/workspaces/ratelimiter/
├── internal/middleware/
│   └── integration_test.go      # Testes do middleware
├── test/
│   ├── memory.go               # Storage em memória para testes
│   ├── integration_test.go     # Testes de integração completa
│   └── redis_integration_test.go # Testes específicos do Redis
└── TESTING.md                 # Este arquivo
```

## Troubleshooting

### Redis não conecta
- Verificar se Redis está rodando: `redis-cli ping`
- Verificar porta: `netstat -an | grep 6379`
- Verificar logs: `docker logs <redis-container>`

### Testes lentos
- Alguns testes incluem `time.Sleep()` para testar reset de limites
- Use `-short` para pular testes longos: `go test -short ./...`

### Falhas intermitentes
- Testes de timing podem falhar em sistemas lentos
- Ajustar durações nos testes se necessário
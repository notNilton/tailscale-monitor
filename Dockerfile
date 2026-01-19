# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Instala dependências de build
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copia go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copia código fonte
COPY . .

# Build dos binários
RUN CGO_ENABLED=1 GOOS=linux go build -o agent ./cmd/agent
RUN CGO_ENABLED=1 GOOS=linux go build -o client ./cmd/client

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Instala dependências runtime e Tailscale
RUN apk add --no-cache ca-certificates sqlite-libs tailscale

# Copia binários do builder
COPY --from=builder /build/agent /app/agent
COPY --from=builder /build/client /app/client

# Copia configuração de exemplo
COPY config.example.yaml /app/config.example.yaml

# Cria diretório de dados
RUN mkdir -p /data

# Expõe porta
EXPOSE 8080

# Volume para dados persistentes
VOLUME ["/data"]

# Comando padrão
CMD ["/app/agent"]

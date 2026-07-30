# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26-alpine AS build

WORKDIR /src

# Layer cache : dépendances d'abord, code ensuite
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/infra-maps-api ./cmd/server

# ---- Runtime stage ----
# distroless static : pas de shell, pas de package manager, non-root par défaut
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/infra-maps-api /infra-maps-api

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/infra-maps-api"]

# Single static Go binary on a distroless, non-root base. Templates, art,
# words and migrations are embedded, so the image needs nothing else.
FROM golang:1.27.4 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/paradise ./cmd/paradise

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/paradise /paradise
USER nonroot:nonroot
EXPOSE 8080
# Railway health-checks GET /healthz (.railway/railway.ts).
ENTRYPOINT ["/paradise"]

# Stage 1: build frontend
FROM node:20-alpine AS web
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build Go binary with embedded frontend
FROM golang:1.26-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# Stage 3: minimal runtime
FROM alpine:3.20
RUN adduser -D -u 1000 app && mkdir -p /data && chown app:app /data
COPY --from=server /out/server /usr/local/bin/server
USER app
EXPOSE 8080
CMD ["/usr/local/bin/server"]

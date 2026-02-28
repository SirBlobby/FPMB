FROM node:22-alpine AS frontend
WORKDIR /app
COPY package.json bun.lockb* ./
RUN npm install
COPY . .
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fpmb-server ./cmd/api/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app/server
COPY --from=backend /fpmb-server ./fpmb-server
COPY --from=frontend /app/build ../build
COPY --from=frontend /app/static ../static
RUN mkdir -p ../data
EXPOSE 8080
CMD ["./fpmb-server"]

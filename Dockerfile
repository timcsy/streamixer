FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src/ ./src/
RUN CGO_ENABLED=0 GOOS=linux go build -o streamixer ./src

FROM alpine:3.19

RUN apk add --no-cache ffmpeg font-noto-cjk fontconfig

COPY --from=builder /app/streamixer /usr/local/bin/streamixer
COPY static/ /app/static/

WORKDIR /app

RUN mkdir -p /dev/shm/streamixer /media /fonts/user /usr/share/fonts/user

EXPOSE 8080

ENV MEDIA_DIR=/media
ENV TMP_DIR=/dev/shm/streamixer
ENV PORT=8080

CMD ["streamixer"]

FROM golang:1.20-alpine

WORKDIR /app

# 必要な依存関係をインストール
COPY ./src/go.mod ./src/go.sum ./
RUN go mod download

# ソースコードをコピー
COPY ./src /app/src

# アプリケーションをビルド
RUN go build -o /app/bin/myapp /app/src/main.go

CMD ["/app/bin/myapp"]
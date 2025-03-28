FROM golang:1.23-alpine

WORKDIR /app

# 依存関係をコピー
COPY go.mod ./

# ソースコードをコピー
COPY *.go ./

# 依存関係をインストール
RUN go mod tidy
RUN go get github.com/gin-contrib/cors
RUN go get github.com/gin-gonic/gin

# アプリケーションをビルド
RUN go build -o app .

EXPOSE 8080

# 実行
CMD ["./app"]
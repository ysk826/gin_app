FROM golang:1.23-alpine

WORKDIR /app

# airをインストール（ホットリロード用）
RUN go install github.com/air-verse/air@latest

# 依存関係をコピー
COPY go.mod ./

# 依存関係をインストール
RUN go mod tidy
RUN go get github.com/gin-contrib/cors
RUN go get github.com/gin-gonic/gin

# airの設定ファイルをコピー
COPY .air.toml ./

EXPOSE 8080

# airを使って実行（ホットリロード対応）
CMD ["air", "-c", ".air.toml"]
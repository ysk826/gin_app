FROM golang:1.23-alpine

# CGOに必要なパッケージをインストール
RUN apk add --no-cache gcc musl-dev

# タイムゾーンパッケージをインストールして日本時間に設定
RUN apk add --no-cache tzdata
ENV TZ=Asia/Tokyo

# CGOを有効化
ENV CGO_ENABLED=1

WORKDIR /gin_app/backend

# airをインストール（ホットリロード用）
RUN go install github.com/air-verse/air@latest

# 依存関係をコピー
COPY go.mod ./

# 依存関係をインストール
RUN go mod tidy
RUN go get github.com/gin-contrib/cors
RUN go get github.com/gin-gonic/gin
RUN go get github.com/mattn/go-sqlite3

# airの設定ファイルをコピー
COPY .air.toml ./

EXPOSE 8080

# airを使って実行（ホットリロード対応）
CMD ["air", "-c", ".air.toml"]
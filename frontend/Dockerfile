FROM node:18-alpine

WORKDIR /app

# package.jsonとpackage-lock.jsonをコピー
COPY package*.json ./

# 依存関係をインストール
RUN npm install

# ソースコードをコピー
COPY . .

# 必要なディレクトリが存在することを確認
RUN mkdir -p public
RUN mkdir -p src

EXPOSE 3000

CMD ["npm", "start"]
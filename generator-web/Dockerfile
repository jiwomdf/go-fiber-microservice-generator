FROM node:22-alpine AS builder

WORKDIR /app

COPY generator-web/package.json ./
RUN npm install

COPY generator-web/index.html ./
COPY generator-web/tsconfig.json ./
COPY generator-web/src ./src

RUN npm run build

FROM nginx:1.27-alpine

COPY generator-web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8090

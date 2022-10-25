# one
FROM golang:1.17 as one
WORKDIR /src/
ENV GOPROXY=https://proxy.golang.com.cn,direct
COPY go.mod ./
RUN go mod download
COPY . .
RUN make build
# two
FROM debian:10-slim
ENV PORT=8080
WORKDIR /
COPY --from=one /src/dist/httpbin ./
EXPOSE $PORT
ENTRYPOINT ["/httpbin", "serve"]

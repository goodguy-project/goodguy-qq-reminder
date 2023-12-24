FROM golang:1.20-bullseye
RUN go env -w GO111MODULE=on
RUN go env -w GOPROXY=https://goproxy.cn,direct
WORKDIR /home
RUN wget https://github.com/Mrs4s/go-cqhttp/releases/download/v1.0.1/go-cqhttp_linux_amd64.tar.gz
RUN tar -zxvf go-cqhttp_linux_amd64.tar.gz --wildcards 'go-cqhttp'
COPY . /home
RUN go build -o goodguy main.go core.go
RUN go build -o qq qq.go
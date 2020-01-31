FROM golang:alpine
COPY . /src
WORKDIR /src
RUN go build -o app main.go
ENTRYPOINT [ "/src/app" ]
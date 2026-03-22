FROM golang:1.26.0 as builder
WORKDIR /src/dnd_back

COPY go.mod go.sum ./
RUN go mod download

COPY ./api ./api
COPY ./auth ./auth
COPY ./model ./model
COPY ./server ./server
COPY main.go ./

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o dnd_back

FROM scratch
COPY --from=builder /src/dnd_back/dnd_back /server
CMD ["/server"]
EXPOSE 8080
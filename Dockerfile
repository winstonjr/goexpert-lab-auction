FROM golang:1.23 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auction ./cmd/auction/main.go

FROM scratch
WORKDIR /app
COPY --from=build /app/auction .
COPY --from=build /app/cmd/auction/.env .
EXPOSE 8080
ENTRYPOINT ["/app/auction"]

#FROM golang:1.20
#
#WORKDIR /app
#
#COPY go.mod ./
#COPY go.sum ./
#RUN go mod download
#
#COPY . .
#
#RUN go build -o /app/auction cmd/auction/main.go
#
#EXPOSE 8080
#
#ENTRYPOINT ["/app/auction"]
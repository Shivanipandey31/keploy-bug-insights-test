FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /notes-app .

FROM alpine:3.19
RUN apk add --no-cache dumb-init
COPY --from=build /notes-app /notes-app
EXPOSE 8080
ENTRYPOINT ["dumb-init", "--"]
CMD ["/notes-app"]

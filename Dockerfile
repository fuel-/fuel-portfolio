FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /portfolio ./cmd/server

FROM scratch
COPY --from=build /portfolio /portfolio
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/portfolio", "-addr", ":8080", "-db", "/data/inquiries.db"]

FROM golang:1.26-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /portfolio ./cmd/server

FROM scratch
# scratch has no CA roots; without these, outbound HTTPS (Resend) fails to
# verify certs with "x509: certificate signed by unknown authority".
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /portfolio /portfolio
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/portfolio", "-addr", ":8080", "-db", "/data/inquiries.db"]

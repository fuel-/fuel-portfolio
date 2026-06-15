TAILWIND := ./bin/tailwindcss.exe

.PHONY: css dev test build

css:
	$(TAILWIND) -i tailwind.css -o static/css/site.css --minify

dev: css
	go run ./cmd/server -db dev.db

test:
	go test ./...

build: css
	go build -ldflags="-s -w" -o bin/portfolio.exe ./cmd/server

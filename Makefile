.PHONY: dev build test generate

generate:
	templ generate
	npx tailwindcss -i ./static/css/input.css -o ./static/css/app.css

dev:
	air

build: generate
	go build -ldflags "-s -w" -o janitarr ./src

test:
	go test -race ./...

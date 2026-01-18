.PHONY: build test generate nix-build

generate:
	templ generate
	./node_modules/.bin/tailwindcss -i ./static/css/input.css -o ./static/css/app.css

build: generate
	go build -ldflags "-s -w" -o janitarr ./src

test:
	go test -race ./...

nix-build:
	nix build .#app
	@echo "Binary available at: result/bin/janitarr"

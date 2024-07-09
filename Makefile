build:
	npx tailwindcss -i ./assets/input.css -o ./assets/output.css
	templ generate

build-and-run:
	go run ./server/server.go
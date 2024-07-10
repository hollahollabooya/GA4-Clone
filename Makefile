TEMPLATES = $(wildcard ./templates/*.templ)
TEMPL_OUTPUT = $(TEMPLATES:.templ=_templ.go)
TAILWIND_INPUT = ./assets/input.css
TAILWIND_OUTPUT = ./assets/output.css
GO_SOURCE = ./server/server.go

all: templ tailwind

templ: $(TEMPL_OUTPUT)

$(TEMPL_OUTPUT): %_templ.go: %.templ
	templ generate

tailwind: $(TAILWIND_OUTPUT)

$(TAILWIND_OUTPUT): $(TAILWIND_INPUT) $(TEMPLATES)
	npx tailwindcss -i $(TAILWIND_INPUT) -o $(TAILWIND_OUTPUT)

run: all
	go run $(GO_SOURCE)

clean:
	rm -f $(TAILWIND_OUTPUT) $(TEMPL_OUTPUT)

.PHONY: all templ tailwind run clean
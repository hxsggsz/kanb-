.PHONY: build run clean dev

build:
	go build -o kanba .

run: build
	./kanba

dev:
	go run .

clean:
	rm -f kanba
	rm -f debug.log

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

tag:
	if [ -z $$NEW_TAG ]; then echo -e "\033[0;31mErro. Por favor, passe a nova tag para ser gerada. Ex: make tag NEW_TAG=1.0.0\033[0m"; exit -1; fi
	LAST_TAG=$(shell git describe --tags --abbrev=0)
	@echo $$NEW_TAG
	git checkout main
	git fetch origin 
	git pull --rebase
	git tag -a v$$NEW_TAG $$(git rev-parse HEAD) -m "$$(git log $$(git describe --tags --abbrev=0)..HEAD --oneline --no-merges --pretty=format:%s)"
	git push --tags

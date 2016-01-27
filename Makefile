configure:
	gb vendor update --all

build:
	gofmt -w src/fieldextract
	go tool vet src/fieldextract/*.go
	gb test
	gb build
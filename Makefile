build:
	go build -o bin/
build-prod:
	go build -ldflags "-s -w" -o bin/ -tags prod
run:
	go run .
tests:
	go test ./...
clean:
	rm -rf bin/

build-mac: clean
	mkdir bin
	fyne package -os darwin -icon ./Icon.png --release --tags prod
	mv psshclient.app bin/

build-linux: clean
	mkdir bin
	fyne package -os linux -icon ./Icon.png --release --tags prod
	mv psshclient.tar.xz bin/

build-win: clean
	mkdir bin
	fyne package -os windows -icon ./Icon.png --release --tags prod --appID co.ispps.psshclient
	mv psshclient.exe bin/
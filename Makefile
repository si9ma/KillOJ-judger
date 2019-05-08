all: build

build:
	go build -o kjudger

clean:
	rm -rf kjudger

install:
	go build -o $$GOPATH/bin/kjudger

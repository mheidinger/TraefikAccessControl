sourcefiles = $(wildcard **/*.go)

build: $(sourcefiles)
	go build -o TraefikAccessControl ./cmd/TraefikAccessControl

run: build
	./TraefikAccessControl

test:
	go test ./...

cleardb:
	-rm tac.db

clearall:	cleardb
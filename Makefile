sourcefiles = $(wildcard **/*.go)

build: $(sourcefiles)
	go build -o TraefikAccessControl ./cmd/TraefikAccessControl

run: build
	./TraefikAccessControl

run-import: build gen-data
	./TraefikAccessControl -import_name tac_data.json -force_import

gen-data:
	go run testData.go

test:
	go test ./...

clean:
	-rm TraefikAccessControl
	-rm tac.db
	-rm tac_data.json
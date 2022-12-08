build:
	mkdir -p target
	GOOS=linux   GOARCH=amd64 go build -o target/main     -ldflags="-w -s" main.go
	GOOS=windows GOARCH=amd64 go build -o target/main.exe -ldflags="-w -s" main.go
	ls -lh target/main target/main.exe

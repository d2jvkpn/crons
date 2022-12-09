build:
	mkdir -p target
	GOOS=linux   GOARCH=amd64 go build -o target/crons     -ldflags="-w -s" main.go
	GOOS=windows GOARCH=amd64 go build -o target/crons.exe -ldflags="-w -s" main.go
	ls -lh target/*

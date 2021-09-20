build:
	go build -o req-forw *.go

build-win:
	GOOS=windows GOARCH=amd64 go build -o req-forw.exe *.go

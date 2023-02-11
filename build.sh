#LINUX BUILD
GOOS=linux GOARCH=amd64 go build -o build/proximal src/main.go
GOOS=linux GOARCH=386 go build -o build/proximal32 src/main.go

#WINDOWS BUILD
GOOS=windows GOARCH=amd64 go build -o build/proximal.exe src/main.go
GOOS=windows GOARCH=386 go build -o build/proximal32.exe src/main.go

#MACOS BUILD
GOOS=darwin GOARCH=amd64 go build -o build/mac-proximal src/main.go


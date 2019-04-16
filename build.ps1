
New-Item -ItemType Directory -Force -Path bin

go build -a -o bin/wait-host.exe cmd/main.go
set GOARCH=amd64

set GOOS=linux
go build -ldflags "-s -w" -trimpath -o release/gost_linux_x64 cmd/gost/main.go cmd/gost/cfg.go cmd/gost/peer.go cmd/gost/route.go

set GOOS=windows
go build -ldflags "-s -w" -trimpath -o release/gost_windows_x64.exe cmd/gost/main.go cmd/gost/cfg.go cmd/gost/peer.go cmd/gost/route.go

set GOOS=darwin
go build -ldflags "-s -w" -trimpath -o release/gost_macos_x64 cmd/gost/main.go cmd/gost/cfg.go cmd/gost/peer.go cmd/gost/route.go

set GOARCH=arm64

set GOOS=linux
go build -ldflags "-s -w" -trimpath -o release/gost_linux_arm64 cmd/gost/main.go cmd/gost/cfg.go cmd/gost/peer.go cmd/gost/route.go

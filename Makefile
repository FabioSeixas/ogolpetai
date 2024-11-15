compile:
	# compile for linux
	GOOS=linux GOARCH=amd64 go build -o ./bin/ogolpetai_linux_amd64 ./cmd/ogolpetai/main.go
	# compile for macOS
	GOOS=darwin GOARCH=amd64 go build -o ./bin/ogolpetai_darwin_amd ./cmd/ogolpetai/main.go
	# compile for apple M1
	GOOS=darwin GOARCH=arm64 go build -o ./bin/ogolpetai_darwin_arm ./cmd/ogolpetai/main.go
	# compile for Windows
	GOOS=windows GOARCH=amd64 go build -o ./bin/ogolpetai_win_amd64 ./cmd/ogolpetai/main.go

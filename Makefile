run:
	xcopy /E /I .\asset ".\build\asset"
	go build -o build/main ./cmd/server
	./build/main
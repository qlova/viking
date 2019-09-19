 all:
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o playground.wasm
	./server

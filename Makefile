server:
	go build -o ./bin/xch-server ./src/server/

s:
	make server

client:
	go build -o ./bin/xch-client ./src/client

c:
	make client

build:
	make server
	make client

b:
	make build

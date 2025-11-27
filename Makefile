server:
	go build -o ./bin/xch-server ./server

s:
	make server

client:
	go build -o ./bin/xch-client ./client

c:
	make client

build:
	make server
	make client

b:
	make build

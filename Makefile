release=creds

build:
	env GOOS=linux GOARCH=amd64 go build -o .build/linux/$(release) ./cmd/$(release)/	

install: build	
	if [ $(shell systemctl show --user --quiet -p ActiveState --value creds) = active ]; then\
		systemctl --user stop creds;\
	fi
	sudo .build/linux/$(release) installsudo
	.build/linux/$(release) installuser

run:
	go run ./cmd/creds/
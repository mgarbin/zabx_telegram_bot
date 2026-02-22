build:
	go build -o zabbix-telegram-notifier main.go

test:
	go test ./...
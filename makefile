dev:
	go run cmd/api/main.go --stage=dev

devJQ:
	go run cmd/api/main.go --stage=dev | jq

prod: 
	go run cmd/api/main.go --stage=prod
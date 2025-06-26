dev:
	go run -race cmd/api/main.go --stage=dev --private --example-data

devJQ:
	go run -race cmd/api/main.go --stage=dev --private --example-data | jq

prod: 
	go run cmd/api/main.go --stage=prod
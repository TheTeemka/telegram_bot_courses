dev:
	go run -race main.go --stage=dev --private --example-data

devJQ:
	go run -race main.go --stage=dev --private --example-data | jq

prod: 
	go run main.go --stage=prod
dev:
	go run -race main.go --stage=dev --private --example-data

devJQ:
	go run -race main.go --stage=dev --private --example-data | jq

prod: 
	go run main.go --stage=prod 

dock_image:
	docker build -t telegram-bot-cources .

dock_run:
	docker run -v ./data:/app/data -d --name telegram-bot-cources telegram-bot-cources ./telegram-bot --private --example-data
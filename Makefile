gen-proto:
	protoc apipb/apis.proto --go_out=plugins=grpc:..

run-server:
	go run server/apisserver.go
	
run-client:
	go run client/apisclient.go
	
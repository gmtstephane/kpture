
all : proto

clean: cleanproto

proto:
	protoc  -I./ kpture/*.proto  --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative 

cleanproto:
	rm ./kpture/*.pb.go

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
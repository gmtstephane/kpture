
all : proto

clean: cleanproto

proto:
	protoc  -I./ capture/*.proto  --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative 

cleanproto:
	rm ./capture/*.pb.go

deps:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
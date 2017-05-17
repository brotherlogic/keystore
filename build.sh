protoc --proto_path ../../../ -I=./testproto --go_out=plugins=grpc:./testproto testproto/test.proto
protoc --proto_path ../../../ -I=./proto --go_out=plugins=grpc:./proto proto/server.proto

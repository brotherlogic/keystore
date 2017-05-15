protoc --proto_path ../../../ -I=./testproto --go_out=plugins=grpc:./testproto testproto/test.proto

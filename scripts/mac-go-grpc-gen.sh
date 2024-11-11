brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

echo export GO_PATH=~go
echo export PATH=$PATH:$GO_PATH/bin

source ~/.zshrc
source ~/.bash_profile
source ~/.bashrc
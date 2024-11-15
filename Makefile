.PHONY: fmt lint test testacc docker test_s3_groups test_crud_group test_user_import test_group_import test_crud_user test_groups test_envs generate install_dnf

# Careful -> no empty string at the end of line
.EXPORT_ALL_VARIABLES:
GOBIN=/home/bin/
TF_CLI_CONFIG_FILE=/home/.terraformrc
TF_LOG=debug
TF_ACC=1

docker: 
	docker build -f Dockerfile -t storagegrid_dev:latest .

install_dnf:
	dnf install -y golang
	go install -v golang.org/x/tools/gopls@latest
	go install -v golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

lint: fmt
	golangci-lint run internal/provider

build: 
	go mod tidy; go install .

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e internal
	gofmt -s -w -e tests/test.go
	terraform fmt -write=true -recursive=true tests/terraform 

testacc:
	go test -v -cover -timeout 120m ./...

test :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/provider_userpass plan -var-file=../variables.tfvars

test_groups :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/data_users_groups init;\
	terraform -chdir=tests/terraform/data_users_groups plan

test_s3_groups :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/data_s3 init;\
	terraform -chdir=tests/terraform/data_s3 plan;\
	terraform -chdir=tests/terraform/data_s3 apply -auto-approve

test_envs :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/provider_envvars plan -var-file=../variables.tfvars
	
test_crud_group :
	rm -rf bin/terraform-provider-storagegrid
	# rm -rf tests/terraform/crud_groups/terraform.tfstate
	# rm -rf tests/terraform/crud_groups/terraform.tfstate.backup
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/crud_groups init;\
	terraform -chdir=tests/terraform/crud_groups plan;\
	terraform -chdir=tests/terraform/crud_groups apply -auto-approve #;\
	#terraform -chdir=tests/terraform/crud_groups destroy -auto-approve

test_group_import :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/crud_groups init;\
	terraform -chdir=tests/terraform/crud_groups state rm storagegrid_groups.new-local-group;\
	terraform -chdir=tests/terraform/crud_groups import storagegrid_groups.new-local-group a88ed298-06d7-4381-b032-a1d28d3eb1dd

test_crud_user :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/crud_users init;\
	terraform -chdir=tests/terraform/crud_users plan;\
	terraform -chdir=tests/terraform/crud_users apply -auto-approve #;\
	#terraform -chdir=tests/terraform/crud_users destroy -auto-approve

test_user_import :
	rm -rf bin/terraform-provider-storagegrid
	go mod tidy; go install .;\
	terraform -chdir=tests/terraform/crud_users init;\
	# terraform -chdir=tests/terraform/crud_users state rm storagegrid_users.new-local-user;\
	terraform -chdir=tests/terraform/crud_users import storagegrid_users.new-local-user b0789794-8aab-4308-985b-55ea4987e91b


.PHONY: test build
test:
	go test -json -timeout 30s ./pkg/kpture -kubeconfig=/Users/stephaneguillemot/.kube/k3s  | gotestfmt
	go test -json -timeout 30s ./test/e2e/... -kubeconfig=/Users/stephaneguillemot/.kube/k3s  | gotestfmt
build:
	mkdir -p ./bin
	go build -o ./bin/agent cmd/agent/main.go 
	go build -o ./bin/proxy cmd/proxy/main.go 

docker:
	docker build  -t ghcr.io/gmtstephane/kpture_proxy:latest . -f Dockerfile.proxy 
	docker push ghcr.io/gmtstephane/kpture_proxy:latest
	docker build  -t ghcr.io/gmtstephane/kpture:latest . -f Dockerfile.agent
	docker push ghcr.io/gmtstephane/kpture:latest

buildx_proxy:
	buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/gmtstephane/kpture_proxy:latest . -f Dockerfile.proxy --push
buildx_agent:
	buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/gmtstephane/kpture:latest . -f Dockerfile.agent --push

buildx: buildx_proxy buildx_agent

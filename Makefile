.PHONY: test build docs

## DOCKER  ##
# build multi plateform proxy and push                             
buildx_proxy:
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/gmtstephane/kpture_proxy:latest . --build-arg BUILDTAG=proxy  --push

# build multi plateform agent and push 
buildx_agent:
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/gmtstephane/kpture:latest . --build-arg BUILDTAG=agent  --push

# build both docker images and push
buildx: buildx_proxy buildx_agent

## CLI ##
# Install kpture with proxy,cli,agent and generate completion script (path must be in $fpath for zsh)   
install :
	go install --tags proxy,cli,agent 
	kpture completion zsh > ~/.oh-my-zsh/completions/_kpture

## RELEASE ##
release:
	goreleaser release --config ./ci/.goreleaser.yaml --snapshot --clean

## Documentation Configuration  ##
docs:
	go run --tags=all main.go docs
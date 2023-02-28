test:
	go test -json -timeout 30s ./pkg/kpture -kubeconfig=/Users/stephaneguillemot/.kube/k3s  | gotestfmt
	go test -json -timeout 30s ./test/e2e/... -kubeconfig=/Users/stephaneguillemot/.kube/k3s  | gotestfmt
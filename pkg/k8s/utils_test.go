package k8s

type (
	kubeProxyHandlerMockGETState       int
	kubeProxyHandlerMockUpdateEphState int
	kubeProxyHandlerMockCREATEState    int
	kubeProxyHandlerMockDELETEState    int
	kubeProxyHandlerMockLISTState      int
)

const (
	getUnkown              kubeProxyHandlerMockGETState = 0
	getOK                  kubeProxyHandlerMockGETState = 1
	getOKNoEph             kubeProxyHandlerMockGETState = 2
	getOkNotReadyThenReady kubeProxyHandlerMockGETState = 3
	getOkNotReadyTimeout   kubeProxyHandlerMockGETState = 4
	getError               kubeProxyHandlerMockGETState = 5
)

const (
	deleteUnkown kubeProxyHandlerMockDELETEState = 0
	deleteOk     kubeProxyHandlerMockDELETEState = 1
	deleteError  kubeProxyHandlerMockDELETEState = 2
)

const (
	createUnkown kubeProxyHandlerMockCREATEState = 0
	createOK     kubeProxyHandlerMockCREATEState = 1
	createError  kubeProxyHandlerMockCREATEState = 2
)

const (
	listUnkown kubeProxyHandlerMockLISTState = 0
	listOK     kubeProxyHandlerMockLISTState = 1
	listError  kubeProxyHandlerMockLISTState = 2
)

const (
	updateUnkown   kubeProxyHandlerMockUpdateEphState = 0
	updateEphOK    kubeProxyHandlerMockUpdateEphState = 1
	updateEphError kubeProxyHandlerMockUpdateEphState = 2
)

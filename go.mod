module github.com/oneinfra/oneinfra

go 1.14

require (
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.1.1
	go.etcd.io/etcd v0.5.0-alpha.5.0.20200819165624-17cef6e3e9d5
	go.uber.org/zap v1.10.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/grpc v1.27.0
	k8s.io/api v0.19.0
	k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery v0.19.0
	k8s.io/client-go v0.19.0
	k8s.io/cluster-bootstrap v0.19.0
	k8s.io/cri-api v0.19.0
	k8s.io/klog/v2 v2.2.0
	k8s.io/kubelet v0.19.0
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/go-logr/logr => github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.2.0
)

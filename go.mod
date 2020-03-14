module github.com/oneinfra/oneinfra

go 1.13

require (
	github.com/go-logr/logr v0.1.0
	github.com/pkg/errors v0.8.1
	github.com/urfave/cli/v2 v2.1.1
	go.etcd.io/etcd v0.5.0-alpha.5.0.20191023171146-3cf2f69b5738
	go.uber.org/zap v1.10.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20200205215550-e35592f146e4
	google.golang.org/grpc v1.23.1
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
	k8s.io/cluster-bootstrap v0.17.3
	k8s.io/cri-api v0.17.2
	k8s.io/klog v1.0.0
	k8s.io/kubelet v0.17.4
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
)

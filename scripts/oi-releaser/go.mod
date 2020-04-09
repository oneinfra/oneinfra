module github.com/oneinfra/oneinfra/scripts/oi-releaser

go 1.14

require (
	github.com/oneinfra/oneinfra v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/urfave/cli/v2 v2.2.0
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace github.com/oneinfra/oneinfra => ../../

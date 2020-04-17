/*
Copyright 2020 Rafael Fernández López <ereslibre@ereslibre.es>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clusterv1alpha1 "github.com/oneinfra/oneinfra/apis/cluster/v1alpha1"
	infrav1alpha1 "github.com/oneinfra/oneinfra/apis/infra/v1alpha1"
	nodev1alpha1 "github.com/oneinfra/oneinfra/apis/node/v1alpha1"
	"github.com/oneinfra/oneinfra/controllers"
	"github.com/oneinfra/oneinfra/internal/pkg/infra"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = infrav1alpha1.AddToScheme(scheme)
	_ = clusterv1alpha1.AddToScheme(scheme)
	_ = nodev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var verbosityLevel int
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&verbosityLevel, "verbosity", 1, "The verbosity level for the controller manager.")
	flag.Set("alsologtostderr", "true")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	klogFlags.Set("v", strconv.Itoa(verbosityLevel))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
		NewClient:          rawClient,
	})
	if err != nil {
		klog.Error("could not set up controller manager")
		os.Exit(1)
	}

	if err = (&controllers.ComponentScheduler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Error("could not set component scheduler controller")
		os.Exit(1)
	}
	if err = (&controllers.ComponentReconciler{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		ConnectionPool: infra.HypervisorConnectionPool{},
	}).SetupWithManager(mgr); err != nil {
		klog.Error("could not set component reconciler controller")
		os.Exit(1)
	}
	if err = (&controllers.ClusterInitializer{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Error("could not set cluster initializer controller")
		os.Exit(1)
	}
	if err = (&controllers.ClusterController{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Error("could not set cluster controller controller")
		os.Exit(1)
	}
	if err = (&controllers.ClusterReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		klog.Error("could not set cluster reconciler controller")
		os.Exit(1)
	}
	if err = (&clusterv1alpha1.Cluster{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Error("could not set up cluster webhook")
		os.Exit(1)
	}
	if err = (&clusterv1alpha1.Component{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Error("could not set up component webhook")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Error("error starting controller manager")
		os.Exit(1)
	}
}

func rawClient(_ cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	return client.New(config, options)
}

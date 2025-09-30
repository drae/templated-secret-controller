// Copyright 2024 The Templatedsecret Controller Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	tsv1alpha1 "github.com/drae/templated-secret-controller/pkg/apis/templatedsecret/v1alpha1"
	"github.com/drae/templated-secret-controller/pkg/generator"
	"github.com/drae/templated-secret-controller/pkg/satoken"
	"github.com/drae/templated-secret-controller/pkg/tracker"

	"github.com/go-logr/logr"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	// Version of templated-secret-controller is set via ldflags at build-time from most recent git tag
	Version       = "develop"
	VersionSuffix = "-dev"
	Commit        = "unknown"

	log                         = logf.Log.WithName("ts")
	ctrlNamespace               = ""
	watchNamespaces             = ""
	metricsBindAddress          = ""
	healthProbeBindAddress      = ""
	enableLeaderElection        = false
	leaderElectionResourceName  = "templated-secret-controller-leader-election"
	reconciliationInterval      = time.Hour
	maxSecretAge                = 720 * time.Hour
	logLevel                    = "info"
	enableCrossNamespaceSecrets = false
	warnOnUnwatchedNamespaces   = true
)

func main() {
	flag.StringVar(&ctrlNamespace, "namespace", "", "Namespace to watch (deprecated, use --watch-namespaces instead)")
	flag.StringVar(&watchNamespaces, "watch-namespaces", "", "Comma-separated list of namespaces to watch (empty for all)")
	flag.StringVar(&metricsBindAddress, "metrics-bind-address", ":8080", "Address for metrics server. If 0, then metrics server doesn't listen on any port.")
	flag.StringVar(&healthProbeBindAddress, "health-probe-bind-address", ":8081", "Address for health probe server (liveness/readiness). If empty or 0, probes are disabled.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager")
	flag.StringVar(&leaderElectionResourceName, "leader-election-id", "templated-secret-controller-leader-election", "Resource name for leader election")
	flag.DurationVar(&reconciliationInterval, "reconciliation-interval", time.Hour, "How often to reconcile SecretTemplates")
	flag.DurationVar(&maxSecretAge, "max-secret-age", 720*time.Hour, "Maximum age of a secret before forcing regeneration")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&enableCrossNamespaceSecrets, "enable-cross-namespace-secret-inputs", false, "Enable experimental cross-namespace Secret inputs (requires source Secret export annotation)")
	flag.BoolVar(&warnOnUnwatchedNamespaces, "warn-on-unwatched-cross-namespaces", true, "Emit warnings when a referenced cross-namespace Secret's namespace is not part of the watch set (may prevent update events)")
	flag.Parse()

	// Set up zap logger with configured log level
	opts := zap.Options{
		Development: false,
	}

	// Configure log level
	switch strings.ToLower(logLevel) {
	case "debug":
		opts.Development = true
	case "info":
		// Default level
	case "warn", "warning":
		// Note: zap Options doesn't have a direct way to set log level through Options
		// This would need a custom zapcore level configuration
	case "error":
		// Note: zap Options doesn't have a direct way to set log level through Options
		// This would need a custom zapcore level configuration
	default:
		fmt.Fprintf(os.Stderr, "Unsupported log level: %s. Using 'info'\n", logLevel)
	}

	logf.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	entryLog := log.WithName("entrypoint")
	entryLog.Info("templated-secret-controller", "version", Version)

	entryLog.Info("setting up manager")
	restConfig := config.GetConfigOrDie()

	// Register API types
	tsv1alpha1.AddToScheme(scheme.Scheme)

	// Wait for CRDs to be ready before starting controller
	entryLog.Info("waiting for SecretTemplate CRD to be ready")
	exitIfErr(entryLog, "waiting for CRDs", waitForCRDs(restConfig, entryLog))

	// Setup manager options
	recoverPanic := true

	// Handle namespace configuration - support both legacy and new approach
	namespaces := make(map[string]cache.Config)
	if watchNamespaces != "" {
		// New approach: watch multiple namespaces
		for _, ns := range strings.Split(watchNamespaces, ",") {
			if ns = strings.TrimSpace(ns); ns != "" {
				namespaces[ns] = cache.Config{}
			}
		}
	} else if ctrlNamespace != "" {
		// Legacy approach: single namespace
		namespaces[ctrlNamespace] = cache.Config{}
	}

	managerOptions := manager.Options{
		// Use proper namespace selector field in newer controller-runtime
		Cache: cache.Options{
			DefaultNamespaces: namespaces,
		},
		Metrics: server.Options{
			BindAddress: metricsBindAddress,
		},
		HealthProbeBindAddress: healthProbeBindAddress,
		// Configure leader election
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        leaderElectionResourceName,
		LeaderElectionNamespace: "", // Use controller namespace if empty
	}

	// Add controller-specific options for newer versions of controller-runtime
	managerOptions.Controller.RecoverPanic = &recoverPanic

	mgr, err := manager.New(restConfig, managerOptions)
	exitIfErr(entryLog, "unable to set up controller manager", err)

	entryLog.Info("setting up controller")

	coreClient, err := kubernetes.NewForConfig(restConfig)
	exitIfErr(entryLog, "building core client", err)

	saLoader := generator.NewServiceAccountLoader(satoken.NewManager(coreClient, log.WithName("template")))

	// Set SecretTemplate's maximum exponential to reduce reconcile time for inputresource errors
	rateLimiter := workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](100*time.Millisecond, 120*time.Second)
	secretTemplateReconciler := generator.NewSecretTemplateReconciler(mgr, mgr.GetClient(), saLoader, tracker.NewTracker(), log.WithName("template"))
	secretTemplateReconciler.SetCrossNamespaceConfig(generator.CrossNamespaceConfig{
		Enabled:           enableCrossNamespaceSecrets,
		WarnOnUnwatched:   warnOnUnwatchedNamespaces,
		WatchedNamespaces: watchedNamespaceSet(namespaces),
	})

	// Pass reconciliation settings to the reconciler
	secretTemplateReconciler.SetReconciliationSettings(reconciliationInterval, maxSecretAge)
	entryLog.Info("configured reconciliation settings",
		"interval", reconciliationInterval.String(),
		"maxSecretAge", maxSecretAge.String())

	exitIfErr(entryLog, "registering", registerCtrlWithRateLimiter("template", mgr, secretTemplateReconciler, rateLimiter))

	entryLog.Info("starting manager")

	// Health check: basic ping
	exitIfErr(entryLog, "adding health check", mgr.AddHealthzCheck("healthz", healthz.Ping))

	// Readiness check: only ready once caches have synced
	cacheSynced := false
	// After start, we will flip this boolean once mgr.GetCache().WaitForCacheSync returns
	readyCheck := func(req *http.Request) error {
		if !cacheSynced {
			return fmt.Errorf("caches not yet synced")
		}
		return nil
	}
	exitIfErr(entryLog, "adding ready check", mgr.AddReadyzCheck("readyz", readyCheck))

	go func() {
		// Wait for the manager to start and caches to sync
		<-mgr.Elected()
		entryLog.Info("waiting for informer caches to sync for readiness")
		if ok := mgr.GetCache().WaitForCacheSync(context.Background()); !ok {
			entryLog.Error(nil, "cache sync failed")
			return
		}
		cacheSynced = true
		entryLog.Info("cache sync complete; controller is ready")
	}()

	err = mgr.Start(signals.SetupSignalHandler())
	exitIfErr(entryLog, "unable to run manager", err)
}

type reconcilerWithWatches interface {
	reconcile.Reconciler
	AttachWatches(controller.Controller) error
}

func registerCtrlWithRateLimiter(desc string, mgr manager.Manager, reconciler reconcilerWithWatches, rateLimiter workqueue.TypedRateLimiter[reconcile.Request]) error {
	ctrlName := "ts-" + desc

	ctrlOpts := controller.Options{
		Reconciler: reconciler,
		// Default MaxConcurrentReconciles is 1. Keeping at that
		// since we are not doing anything that we need to parallelize for.
		RateLimiter: rateLimiter,
	}

	ctrl, err := controller.New(ctrlName, mgr, ctrlOpts)
	if err != nil {
		return fmt.Errorf("%s: unable to set up: %s", ctrlName, err)
	}

	err = reconciler.AttachWatches(ctrl)
	if err != nil {
		return fmt.Errorf("%s: unable to attach watches: %s", ctrlName, err)
	}

	return nil
}

// watchedNamespaceSet converts the controller-runtime namespace config map into a simple set for quick membership tests.
func watchedNamespaceSet(cfg map[string]cache.Config) map[string]struct{} {
	if len(cfg) == 0 { // cluster-wide
		return map[string]struct{}{}
	}
	set := make(map[string]struct{}, len(cfg))
	for ns := range cfg {
		set[ns] = struct{}{}
	}
	return set
}

func exitIfErr(entryLog logr.Logger, desc string, err error) {
	if err != nil {
		entryLog.Error(err, desc)
		os.Exit(1)
	}
}

func waitForCRDs(restConfig *rest.Config, entryLog logr.Logger) error {
	apiExtClient, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("unable to create API extensions client: %w", err)
	}

	// First check that the CRD exists and is established
	err = wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
		crd, err := apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "secrettemplates.templatedsecret.starstreak.dev", metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				entryLog.Info("SecretTemplate CRD not found, retrying...")
				return false, nil
			}
			return false, err
		}

		// Check that the CRD is established
		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextensionsv1.Established &&
				condition.Status == apiextensionsv1.ConditionTrue {
				entryLog.Info("SecretTemplate CRD is established")
				return true, nil
			}
		}

		entryLog.Info("SecretTemplate CRD found but not yet established, retrying...")
		return false, nil
	})

	if err != nil {
		return err
	}

	// Now verify that the API resource is actually discoverable
	// This ensures the apiserver's discovery cache has been updated
	discoveryClient := apiExtClient.Discovery()
	entryLog.Info("Verifying API resource is discoverable")

	return wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		resourceList, err := discoveryClient.ServerResourcesForGroupVersion("templatedsecret.starstreak.dev/v1alpha1")
		if err != nil {
			entryLog.Info("API resource not yet discoverable, waiting for API server discovery cache to refresh...",
				"error", err.Error())
			return false, nil
		}

		for _, r := range resourceList.APIResources {
			if r.Kind == "SecretTemplate" {
				entryLog.Info("SecretTemplate API resource is now discoverable")
				return true, nil
			}
		}

		entryLog.Info("SecretTemplate kind not found in API resources, waiting...")
		return false, nil
	})
}

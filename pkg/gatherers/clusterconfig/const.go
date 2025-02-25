package clusterconfig

import (
	registryv1 "github.com/openshift/api/imageregistry/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	registryScheme = runtime.NewScheme()
	// logMaxLines sets maximum number of lines of the log file
	logMaxTailLines = int64(100)
	// logMaxLongTailLines sets maximum number of lines of the long log file
	logMaxLongTailLines = int64(2000)
	// logLinesOffset sets the maximum offset if a stacktrace message was found in the logs
	logLinesOffset              = 20
	logStackTraceMaxLines       = 40
	logStackTraceBeginningLimit = 35
	logStackTraceEndLimit       = 5
	// logNodeUnit sets the journal unit to be used to collect the node logs (options: kubelet or crio)
	logNodeUnit = "kubelet"
	// logNodeMaxTailLines sets the maximum number of lines to be fetched
	logNodeMaxTailLines = 1000
	// logNodeMaxLines sets the maximum number of lines of the node log to be stored per node
	logNodeMaxLines = 50

	// logDefaultSinceSeconds is used in calls to common.CollectLogsFromContainers
	logDefaultSinceSeconds = int64(86400) // last day
	// logDefaultLimitBytes is used in calls to common.CollectLogsFromContainers
	logDefaultLimitBytes = int64(1024 * 64) // maximum 64 kb of logs

	defaultNamespaces           = []string{"default", "kube-system", "kube-public", "openshift"}
	datahubGroupVersionResource = schema.GroupVersionResource{
		Group: "installers.datahub.sap.com", Version: "v1alpha1", Resource: "datahubs",
	}
	pnccGroupVersionResource = schema.GroupVersionResource{
		Group: "controlplane.operator.openshift.io", Version: "v1alpha1", Resource: "podnetworkconnectivitychecks",
	}
	machineConfigGroupVersionResource = schema.GroupVersionResource{
		Group: "machineconfiguration.openshift.io", Version: "v1", Resource: "machineconfigs",
	}
	machineHeatlhCheckGVR = schema.GroupVersionResource{
		Group: "machine.openshift.io", Version: "v1beta1", Resource: "machinehealthchecks",
	}
	machineAutoScalerGvr = schema.GroupVersionResource{
		Group: "autoscaling.openshift.io", Version: "v1beta1", Resource: "machineautoscalers",
	}
	openshiftLoggingResource = schema.GroupVersionResource{
		Group: "logging.openshift.io", Version: "v1", Resource: "clusterloggings",
	}

	jaegerResource = schema.GroupVersionResource{
		Group: "jaegertracing.io", Version: "v1", Resource: "jaegers",
	}

	costManagementMetricsConfigResource = schema.GroupVersionResource{
		Group: "costmanagement-metrics-cfg.openshift.io", Version: "v1beta1", Resource: "costmanagementmetricsconfigs",
	}
)

func init() { //nolint: gochecknoinits
	utilruntime.Must(registryv1.AddToScheme(registryScheme))
}

package controller

import (
	"context"
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/openshift/insights-operator/pkg/anonymization"
	"github.com/openshift/insights-operator/pkg/authorizer/clusterauthorizer"
	"github.com/openshift/insights-operator/pkg/config"
	"github.com/openshift/insights-operator/pkg/config/configobserver"
	"github.com/openshift/insights-operator/pkg/gather"
	"github.com/openshift/insights-operator/pkg/insights/insightsclient"
	"github.com/openshift/insights-operator/pkg/recorder"
	"github.com/openshift/insights-operator/pkg/recorder/diskrecorder"
)

// GatherJob is the type responsible for controlling a non-periodic Gather execution
type GatherJob struct {
	config.Controller
}

// Gather runs a single gather and stores the generated archive, without uploading it.
// 1. Creates the necessary configs/clients
// 2. Creates the configobserver to get more configs
// 3. Initiates the recorder
// 4. Executes a Gather
// 5. Flushes the results
func (d *GatherJob) Gather(ctx context.Context, kubeConfig, protoKubeConfig *rest.Config) error {
	klog.Infof("Starting insights-operator %s", version.Get().String())
	// these are operator clients
	kubeClient, err := kubernetes.NewForConfig(protoKubeConfig)
	if err != nil {
		return err
	}

	gatherProtoKubeConfig, gatherKubeConfig, metricsGatherKubeConfig, alertsGatherKubeConfig := prepareGatherConfigs(
		protoKubeConfig, kubeConfig, d.Impersonate,
	)

	// ensure the insight snapshot directory exists
	if _, err = os.Stat(d.StoragePath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(d.StoragePath, 0777); err != nil {
			return fmt.Errorf("can't create --path: %v", err)
		}
	}

	// configobserver synthesizes all config into the status reporter controller
	configObserver := configobserver.New(d.Controller, kubeClient)

	var anonymizer *anonymization.Anonymizer
	if anonymization.IsObfuscationEnabled(configObserver) {
		// anonymizer is responsible for anonymizing sensitive data, it can be configured to disable specific anonymization
		anonymizer, err = anonymization.NewAnonymizerFromConfig(ctx, gatherKubeConfig, gatherProtoKubeConfig, protoKubeConfig)
		if err != nil {
			return err
		}
	}

	// the recorder stores the collected data and we flush at the end.
	recdriver := diskrecorder.New(d.StoragePath)
	rec := recorder.New(recdriver, d.Interval, anonymizer)
	defer func() {
		if err := rec.Flush(); err != nil {
			klog.Error(err)
		}
	}()

	authorizer := clusterauthorizer.New(configObserver)
	insightsClient := insightsclient.New(nil, 0, "default", authorizer, gatherKubeConfig)

	gatherers := gather.CreateAllGatherers(
		gatherKubeConfig, gatherProtoKubeConfig, metricsGatherKubeConfig, alertsGatherKubeConfig, anonymizer,
		configObserver, insightsClient,
	)

	allFunctionReports := make(map[string]gather.GathererFunctionReport)
	for _, gatherer := range gatherers {
		functionReports, err := gather.CollectAndRecordGatherer(ctx, gatherer, rec, configObserver)
		if err != nil {
			klog.Errorf("unable to process gatherer %v, error: %v", gatherer.GetName(), err)
			functionReports = append(functionReports, gather.GathererFunctionReport{
				FuncName: gatherer.GetName(),
				Errors:   []string{err.Error()},
			})
		}

		for i := range functionReports {
			allFunctionReports[functionReports[i].FuncName] = functionReports[i]
		}
	}

	return gather.RecordArchiveMetadata(mapToArray(allFunctionReports), rec, anonymizer)
}

func mapToArray(m map[string]gather.GathererFunctionReport) []gather.GathererFunctionReport {
	a := make([]gather.GathererFunctionReport, 0, len(m))
	for _, v := range m {
		a = append(a, v)
	}
	return a
}

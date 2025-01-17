package resource

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	chaosTypes "github.com/litmuschaos/chaos-operator/pkg/types"
)

var (
	gvrdc = schema.GroupVersionResource{
		Group:    "apps.openshift.io",
		Version:  "v1",
		Resource: "deploymentconfigs",
	}
)

// CheckDeploymentConfigAnnotation will check the annotation of deployment
func CheckDeploymentConfigAnnotation(clientSet dynamic.Interface, engine *chaosTypes.EngineInfo) (*chaosTypes.EngineInfo, error) {
	deploymentConfigList, err := getDeploymentConfigList(clientSet, engine)
	if err != nil {
		return engine, err
	}

	engine, chaosEnabledDeploymentConfig, err := checkForChaosEnabledDeploymentConfig(deploymentConfigList, engine)
	if err != nil {
		return engine, err
	}

	if chaosEnabledDeploymentConfig == 0 {
		return engine, errors.New("no deploymentconfigs chaos-candidate found")
	}

	return engine, nil
}

func getDeploymentConfigList(clientSet dynamic.Interface, engine *chaosTypes.EngineInfo) (*unstructured.UnstructuredList, error) {
	dynamicClient := clientSet.Resource(gvrdc)

	deploymentConfigList, err := dynamicClient.Namespace(engine.AppInfo.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: engine.Instance.Spec.Appinfo.Applabel})
	if err != nil {
		return nil, fmt.Errorf("error while listing deploymentconfigs with matching labels: %s", engine.Instance.Spec.Appinfo.Applabel)
	}

	if len(deploymentConfigList.Items) == 0 {
		return nil, fmt.Errorf("no deploymentconfigs with matching labels: %s", engine.Instance.Spec.Appinfo.Applabel)
	}

	return deploymentConfigList, err
}

// checkForChaosEnabledDeploymentConfig will check and count the total chaos enabled application
func checkForChaosEnabledDeploymentConfig(deploymentConfigList *unstructured.UnstructuredList, engine *chaosTypes.EngineInfo) (*chaosTypes.EngineInfo, int, error) {
	chaosEnabledDeploymentConfig := 0
	if deploymentConfigList == nil {
		return engine, chaosEnabledDeploymentConfig, fmt.Errorf("deploymentconfigs is nil")
	}

	for _, deploymentConfig := range deploymentConfigList.Items {
		annotationValue := deploymentConfig.GetAnnotations()[ChaosAnnotationKey]
		if IsChaosEnabled(annotationValue) {
			chaosTypes.Log.Info("chaos candidate of", "kind:", engine.AppInfo.Kind, "appName: ", deploymentConfig.GetName(), "appUUID: ", deploymentConfig.GetUID())
			chaosEnabledDeploymentConfig++
		}
	}

	return engine, chaosEnabledDeploymentConfig, nil
}

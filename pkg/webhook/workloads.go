package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rayanebel/k8spratices-admission-controller/pkg/utils"
	"k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

var workloads = []metav1.GroupVersionResource{
	metav1.GroupVersionResource{Group: "", Version: "apps/v1", Resource: "deployments"},
	metav1.GroupVersionResource{Group: "", Version: "apps/v1", Resource: "daemonsets"},
	metav1.GroupVersionResource{Group: "", Version: "apps/v1", Resource: "replicasets"},
	metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
}

func resourceIsWorkload(resource metav1.GroupVersionResource) bool {
	for _, workload := range workloads {
		if workload == resource {
			return true
		}
	}
	return false
}

func validateWorkloads(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.Info("validating workloads...")
	isValidWorkload := resourceIsWorkload(ar.Request.Resource)
	if !isValidWorkload {
		klog.Errorf("expect resource to be of type %v", workloads)
		return nil
	}

	var raw []byte
	if ar.Request.Operation == v1beta1.Delete {
		raw = ar.Request.OldObject.Raw
	} else {
		raw = ar.Request.Object.Raw
	}

	workloadType := ar.Request.Kind.Kind

	reviewResponse := &v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	switch workloadType {
	case "Deployment":
		workload := appsv1.Deployment{}
		if err := json.Unmarshal(raw, &workload); err != nil {
			klog.Error(err)
			return utils.ToAdmissionResponse(err)
		}
		validateDeployment(workload, reviewResponse)
	case "Pod":
		workload := corev1.Pod{}
		if err := json.Unmarshal(raw, &workload); err != nil {
			klog.Error(err)
			return utils.ToAdmissionResponse(err)
		}
		validatePod(workload, reviewResponse)
	case "DaemonSet":
		workload := appsv1.DaemonSet{}
		if err := json.Unmarshal(raw, &workload); err != nil {
			klog.Error(err)
			return utils.ToAdmissionResponse(err)
		}
		validateDaemonSet(workload, reviewResponse)
	case "ReplicaSet":
		workload := appsv1.ReplicaSet{}
		if err := json.Unmarshal(raw, &workload); err != nil {
			klog.Error(err)
			return utils.ToAdmissionResponse(err)
		}
		validateReplicaSet(workload, reviewResponse)
	}

	return reviewResponse
}

func validateDeployment(resource appsv1.Deployment, response *v1beta1.AdmissionResponse) {
	res := workloadHasResourceProperties(resource.Spec.Template.Spec)
	if !res {
		response.Allowed = false
		message := fmt.Sprintf("Rejecting workload: %v of type %v: one or more container does not have Resource Requests and/or Limits set", resource.Name, resource.Kind)
		response.Result = &metav1.Status{Message: message}
	}
}
func validateReplicaSet(resource appsv1.ReplicaSet, response *v1beta1.AdmissionResponse) {
	res := workloadHasResourceProperties(resource.Spec.Template.Spec)
	if !res {
		response.Allowed = false
		message := fmt.Sprintf("Rejecting workload: %v of type %v: one or more container does not have Resource Requests and/or Limits set", resource.Name, resource.Kind)
		response.Result = &metav1.Status{Message: message}
	}
}
func validateDaemonSet(resource appsv1.DaemonSet, response *v1beta1.AdmissionResponse) {
	res := workloadHasResourceProperties(resource.Spec.Template.Spec)
	if !res {
		response.Allowed = false
		message := fmt.Sprintf("Rejecting workload: %v of type %v: one or more container does not have Resource Requests and/or Limits set", resource.Name, resource.Kind)
		response.Result = &metav1.Status{Message: message}
	}
}
func validatePod(resource corev1.Pod, response *v1beta1.AdmissionResponse) {
	res := workloadHasResourceProperties(resource.Spec)
	if !res {
		response.Allowed = false
		message := fmt.Sprintf("Rejecting workload: %v of type %v: one or more container does not have Resource Requests and/or Limits set", resource.Name, resource.Kind)
		response.Result = &metav1.Status{Message: message}
	}
}

func workloadHasResourceProperties(spec v1.PodSpec) bool {
	for _, container := range spec.Containers {
		if len(container.Resources.Limits) == 0 || len(container.Resources.Requests) == 0 {
			return false
		}
	}
	return true
}

func serveWorkloads(w http.ResponseWriter, r *http.Request) {
	serve(w, r, validateWorkloads)
}

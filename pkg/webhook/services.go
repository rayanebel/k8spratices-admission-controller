package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rayanebel/k8spratices-admission-controller/pkg/utils"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	serviceLoadBalancerType = "LoadBalancer"
	bypassAnnotation        = "security.k8s.thalesdigital.io/allow-no-ip-filtering"
	bypassAnnotationValue   = "true"
)

func validateServices(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	klog.Info("validating services...")
	serviceResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	if ar.Request.Resource != serviceResource {
		klog.Errorf("expect resource to be %s", serviceResource)
		return nil
	}

	var raw []byte
	if ar.Request.Operation == v1beta1.Delete {
		raw = ar.Request.OldObject.Raw
	} else {
		raw = ar.Request.Object.Raw
	}

	service := corev1.Service{}
	if err := json.Unmarshal(raw, &service); err != nil {
		klog.Error(err)
		return utils.ToAdmissionResponse(err)
	}
	reviewResponse := &v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if service.Spec.Type == serviceLoadBalancerType {
		validateServiceLoadBalancer(reviewResponse, service)
	}

	return reviewResponse
}

func validateServiceLoadBalancer(responseObject *v1beta1.AdmissionResponse, service corev1.Service) {
	klog.Infof("validating service: %s", service.Name)
	annotationValue := utils.GetAnnotationValue(service.Annotations, bypassAnnotation)
	if annotationValue == bypassAnnotationValue {
		klog.Infof("Service: %s has bypass annotation: %s", service.Name, bypassAnnotation)
		responseObject.Allowed = true
		return
	}
	if len(service.Spec.LoadBalancerSourceRanges) == 0 {
		responseObject.Allowed = false
		message := fmt.Sprintf("Rejecting service %v of type LoadBalancer: 'spec.loadBalancerSourceRanges' properties must be set", service.Name)
		responseObject.Result = &metav1.Status{Message: message}
		return
	}
}

func serveServices(w http.ResponseWriter, r *http.Request) {
	serve(w, r, validateServices)
}

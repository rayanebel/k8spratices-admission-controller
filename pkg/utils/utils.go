package utils

import (
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

func GetAnnotationValue(annotations map[string]string, annotation string) string {
	for k, v := range annotations {
		if k == annotation {
			return v
		}
	}
	return ""
}

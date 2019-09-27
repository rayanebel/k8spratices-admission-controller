package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/rayanebel/k8spratices-admission-controller/pkg/utils"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/klog"
)

const (
	tlsDir      = `/etc/certs`
	tlsCertFile = `webhook.pem`
	tlsKeyFile  = `webhook-key.pem`
)

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

func InitServer() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	router := mux.NewRouter()
	router.HandleFunc("/services", serveServices)
	router.HandleFunc("/workloads", serveWorkloads)
	server := &http.Server{
		Addr:    ":8443",
		Handler: router,
	}
	klog.Infof("starting webserver on address %s", server.Addr)
	//klog.V(2).Infof("starting webserver on address %s", server.Addr)
	klog.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}

// This function will call the corresponding webhook function and return the response to kubernetes.
func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.V(2).Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	if err := json.Unmarshal(body, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = utils.ToAdmissionResponse(err)
	} else {
		// pass to admitFunc
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

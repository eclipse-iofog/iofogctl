package install

import (
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func initOperatorClientScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	gv := schema.GroupVersion{Group: util.GetCliCrdGroup(), Version: "v3"}
	scheme.AddKnownTypes(gv, &cpv3.ControlPlane{}, &cpv3.ControlPlaneList{})
	metav1.AddToGroupVersion(scheme, gv)

	return scheme
}

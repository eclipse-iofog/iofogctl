package apis

import (
	"github.com/eclipse-iofog/iofog-operator/v2/pkg/apis/iofog"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, iofog.SchemeBuilder.AddToScheme)
}

package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	extsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func controlPlaneCRDName(group string) string {
	return fmt.Sprintf("controlplanes.%s", group)
}

func newControlPlaneCRD(group string) *extsv1.CustomResourceDefinition {
	apiVersions := []string{"v3"}
	versions := make([]extsv1.CustomResourceDefinitionVersion, len(apiVersions))
	preserveUnknownFields := true

	for i, version := range apiVersions {
		versions[i].Name = version
		versions[i].Served = true

		if i == 0 {
			versions[i].Storage = true
		}

		versions[i].Schema = &extsv1.CustomResourceValidation{
			OpenAPIV3Schema: &extsv1.JSONSchemaProps{
				Properties:             map[string]extsv1.JSONSchemaProps{},
				XPreserveUnknownFields: &preserveUnknownFields,
				Type:                   "object",
			},
		}
		versions[i].Subresources = &extsv1.CustomResourceSubresources{
			Status: &extsv1.CustomResourceSubresourceStatus{},
		}
	}

	return &extsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: controlPlaneCRDName(group),
		},
		Spec: extsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: extsv1.CustomResourceDefinitionNames{
				Kind:     "ControlPlane",
				ListKind: "ControlPlaneList",
				Plural:   "controlplanes",
				Singular: "controlplane",
			},
			Scope:    extsv1.NamespaceScoped,
			Versions: versions,
		},
	}
}

func sameCRDVersionsSupported(left, right *extsv1.CustomResourceDefinition) bool {
	for _, leftVersion := range left.Spec.Versions {
		matched := false
		for _, rightVersion := range right.Spec.Versions {
			if leftVersion.Name == rightVersion.Name {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func isSupportedControlPlaneCRD(existing, expected *extsv1.CustomResourceDefinition) bool {
	if existing.Name != expected.Name {
		return false
	}
	return sameCRDVersionsSupported(expected, existing)
}

func controlPlaneCRDsToInstall() []*extsv1.CustomResourceDefinition {
	return []*extsv1.CustomResourceDefinition{newControlPlaneCRD(util.GetCliCrdGroup())}
}

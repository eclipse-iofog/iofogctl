module github.com/eclipse-iofog/iofogctl/v3

go 1.16

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/briandowns/spinner v1.6.1
	github.com/containerd/containerd v1.4.4 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/docker/go-connections v0.4.0
	github.com/eclipse-iofog/iofog-go-sdk/v2 v2.0.0-beta3.0.20210306092845-4d8568558b5d // indirect
	github.com/eclipse-iofog/iofog-go-sdk/v3 v3.0.0-beta1
	github.com/eclipse-iofog/iofog-operator/v3 v3.0.0-beta1
	github.com/eclipse-iofog/iofogctl v1.3.2 // indirect
	github.com/go-logr/logr v0.3.0 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/skupperproject/skupper-cli v0.0.1-beta6 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/twmb/algoimpl v0.0.0-20170717182524-076353e90b94
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.6.4

)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

// iofog-operator
replace (
	github.com/eclipse-iofog/iofog-go-sdk/v3 => github.com/eclipse-iofog/iofog-go-sdk/v3 v3.0.0-20211022021531-d4e44c73af48
	github.com/go-logr/logr => github.com/go-logr/logr v0.3.0
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.3.0
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.1
	k8s.io/client-go => k8s.io/client-go v0.19.4
)

exclude github.com/Sirupsen/logrus v1.4.2

exclude github.com/Sirupsen/logrus v1.4.1

exclude github.com/Sirupsen/logrus v1.4.0

exclude github.com/Sirupsen/logrus v1.3.0

exclude github.com/Sirupsen/logrus v1.2.0

exclude github.com/Sirupsen/logrus v1.1.1

exclude github.com/Sirupsen/logrus v1.1.0

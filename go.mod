module github.com/eclipse-iofog/iofogctl/v3

go 1.15

require (
	cloud.google.com/go v0.51.0 // indirect
	github.com/Azure/go-autorest/autorest v0.9.6 // indirect
	github.com/GeertJohan/go.rice v1.0.0
	github.com/briandowns/spinner v1.6.1
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/docker/go-connections v0.4.0
	github.com/eclipse-iofog/iofog-go-sdk v1.3.0 // indirect
	github.com/eclipse-iofog/iofog-go-sdk/v2 v2.0.0-beta3.0.20210306092845-4d8568558b5d
	github.com/eclipse-iofog/iofog-operator v1.3.0 // indirect
	github.com/eclipse-iofog/iofog-operator/v2 v2.0.0-20210304184121-d30002bc497d // indirect
	github.com/eclipse-iofog/iofog-operator/v3 v3.0.0-20210310234756-435f1376c1da
	github.com/eclipse-iofog/iofogctl v1.3.2
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/imdario/mergo v0.3.10 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/operator-framework/api v0.3.25 // indirect
	github.com/operator-framework/operator-lib v0.2.0 // indirect
	github.com/operator-framework/operator-registry v1.15.3 // indirect
	github.com/operator-framework/operator-sdk v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/skupperproject/skupper-cli v0.0.1-beta6.0.20191022215135-8088454e7fda // indirect
	github.com/spf13/cobra v1.0.0
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/twmb/algoimpl v0.0.0-20170717182524-076353e90b94
	go.uber.org/goleak v1.1.10 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.0.0-20201014231627-1610a49f37af // indirect
	golang.org/x/tools/gopls v0.4.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	helm.sh/helm/v3 v3.4.1 // indirect
	k8s.io/api v0.19.4
	k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubectl v0.19.4 // indirect
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800 // indirect
	sigs.k8s.io/controller-runtime v0.6.4
	sigs.k8s.io/controller-tools v0.4.1 // indirect
	sigs.k8s.io/kubebuilder/v2 v2.3.2-0.20201214213149-0a807f4e9428 // indirect
	sigs.k8s.io/kustomize/kustomize/v3 v3.5.4 // indirect
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

// iofog-operator
replace (
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

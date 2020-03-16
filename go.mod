module github.com/eclipse-iofog/iofogctl/v2

go 1.12

require (
	github.com/GeertJohan/go.rice v1.0.0
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/NYTimes/gziphandler v1.0.1 // indirect
	github.com/Sirupsen/logrus v1.0.6 // indirect
	github.com/briandowns/spinner v1.6.1
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v0.0.0-20180612054059-a9fbbdc8dd87
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/eclipse-iofog/iofog-go-sdk/v2 v2.0.0-alpha.0.20200315200621-20eaaed13c41
	github.com/eclipse-iofog/iofog-operator/v2 v2.0.0-20200316022750-a545fa6ecf0f
	github.com/eclipse-iofog/iofogctl v1.3.2 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/petar/GoLLRB v0.0.0-20190514000832-33fb24c13b99 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/cobra v0.0.5
	github.com/twmb/algoimpl v0.0.0-20170717182524-076353e90b94
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	google.golang.org/grpc v1.22.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.1.10
)

// Pinned to kubernetes-1.13.4
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/docker/docker => github.com/docker/engine v0.0.0-20190717161051-705d9623b7c1
	github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37
	github.com/sirupsen/logrus => github.com/sirupsen/logrus v1.0.1
	k8s.io/api => k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190228174905-79427f02047f
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190228180923-a9e421a79326
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190228175259-3e0149950b0e
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20180711000925-0cf8f7e6ed1d
	k8s.io/kubernetes => k8s.io/kubernetes v1.13.4
)

exclude github.com/Sirupsen/logrus v1.4.2

exclude github.com/Sirupsen/logrus v1.4.1

exclude github.com/Sirupsen/logrus v1.4.0

exclude github.com/Sirupsen/logrus v1.3.0

exclude github.com/Sirupsen/logrus v1.2.0

exclude github.com/Sirupsen/logrus v1.1.1

exclude github.com/Sirupsen/logrus v1.1.0

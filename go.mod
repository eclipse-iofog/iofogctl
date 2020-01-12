module github.com/eclipse-iofog/iofogctl

go 1.12

require (
	cloud.google.com/go v0.39.0 // indirect
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
	github.com/eclipse-iofog/iofog-go-sdk v0.0.0-20200112205444-b0cebcdf3134
	github.com/eclipse-iofog/iofog-operator v0.0.0-20191117223400-e0d2d4ac3fcc
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/json-iterator/go v1.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/petar/GoLLRB v0.0.0-20190514000832-33fb24c13b99 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/spf13/cobra v0.0.4
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9 // indirect
	golang.org/x/oauth2 v0.0.0-20190523182746-aaccbc9213b0 // indirect
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	golang.org/x/text v0.3.2 // indirect
	google.golang.org/appengine v1.6.0 // indirect
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610 // indirect
	google.golang.org/grpc v1.22.1 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20191115135540-bbc9463b57e5
	k8s.io/apiextensions-apiserver v0.0.0-20190228180357-d002e88f6236
	k8s.io/apimachinery v0.0.0-20191115015347-3c7067801da2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20191114200735-6ca3b61696b6 // indirect
	sigs.k8s.io/controller-runtime v0.1.10
)

// Pinned to kubernetes-1.13.4
replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/docker/docker => github.com/docker/engine v0.0.0-20190717161051-705d9623b7c1
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

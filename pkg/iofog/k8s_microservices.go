package iofog

type microservice struct {
	name            string
	port            int
	image           string
	imagePullPolicy string
	replicas        int32
	env             map[string]string
}

var controllerMicroservice = microservice{
	name:            "controller",
	port:            51121,
	image:           "edgeworx/controller-k8s:latest",
	imagePullPolicy: "Always",
	replicas:        1,
}

var connectorMicroservice = microservice{
	name:            "connector",
	port:            8080,
	image:           "iofog/connector:dev",
	imagePullPolicy: "Always",
	replicas:        1,
}

var schedulerMicroservice = microservice{
	name:            "scheduler",
	image:           "iofog/iofog-scheduler:develop",
	imagePullPolicy: "Always",
	replicas:        1,
}

var operatorMicroservice = microservice{
	name:            "connector",
	port:            60000,
	image:           "iofog/iofog-operator:develop",
	imagePullPolicy: "Always",
	replicas:        1,
}

var kubeletMicroservice = microservice{
	name:            "kubelet",
	port:            60000,
	image:           "iofog/iofog-kubelet:develop",
	imagePullPolicy: "Always",
	replicas:        1,
}

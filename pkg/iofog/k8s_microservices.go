package iofog

type microservice struct {
	name       string
	port       int
	replicas   int32
	containers []container
}

type container struct {
	name            string
	image           string
	imagePullPolicy string
	args            []string
}

var controllerMicroservice = microservice{
	name:     "controller",
	port:     51121,
	replicas: 1,
	containers: []container{
		{
			name:            "controller",
			image:           "edgeworx/controller-k8s:latest",
			imagePullPolicy: "Always",
		},
	},
}

var connectorMicroservice = microservice{
	name:     "connector",
	port:     8080,
	replicas: 1,
	containers: []container{
		{
			name:            "connector",
			image:           "iofog/connector:dev",
			imagePullPolicy: "Always",
		},
	},
}

var schedulerMicroservice = microservice{
	name:     "scheduler",
	replicas: 1,
	containers: []container{
		{
			name:            "scheduler",
			image:           "iofog/iofog-scheduler:develop",
			imagePullPolicy: "Always",
		},
	},
}

var operatorMicroservice = microservice{
	name:     "operator",
	port:     60000,
	replicas: 1,
	containers: []container{
		{
			name:            "operator",
			image:           "iofog/iofog-operator:develop",
			imagePullPolicy: "Always",
		},
	},
}

var kubeletMicroservice = microservice{
	name:     "kubelet",
	port:     60000,
	replicas: 1,
	containers: []container{
		{
			name:            "kubelet",
			image:           "iofog/iofog-kubelet:develop",
			imagePullPolicy: "Always",
			args: []string{
				"--namespace",
				"",
				"--iofog-token",
				"",
				"--iofog-url",
				"",
			},
		},
	},
}

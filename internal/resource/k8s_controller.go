package resource

type KubernetesController struct {
	PodName  string `yaml:"podName"`
	Endpoint string `yaml:"endpoint"`
	Created  string `yaml:"created,omitempty"`
	Status   string `yaml:"status,omitempty"`
}

func (ctrl *KubernetesController) GetName() string {
	return ctrl.PodName
}

func (ctrl *KubernetesController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl *KubernetesController) GetCreatedTime() string {
	return ctrl.Created
}

func (ctrl *KubernetesController) SetName(name string) {
	ctrl.PodName = name
}

func (ctrl *KubernetesController) Sanitize() error {
	return nil
}

func (ctrl *KubernetesController) Clone() Controller {
	return &KubernetesController{
		PodName:  ctrl.PodName,
		Endpoint: ctrl.Endpoint,
		Created:  ctrl.Created,
	}
}

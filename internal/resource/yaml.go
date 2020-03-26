package resource

import (
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallKubernetesControlPlane(file []byte) (controlPlane KubernetesControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = controlPlane.Sanitize()
	return
}

func UnmarshallRemoteControlPlane(file []byte) (controlPlane RemoteControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = controlPlane.Sanitize()
	return
}

func UnmarshallLocalControlPlane(file []byte) (controlPlane LocalControlPlane, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controlPlane); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = controlPlane.Sanitize()
	return
}

func UnmarshallRemoteController(file []byte) (controller RemoteController, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controller); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = controller.Sanitize()
	return
}

func UnmarshallLocalController(file []byte) (controller LocalController, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &controller); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = controller.Sanitize()
	return
}

func UnmarshallRemoteAgent(file []byte) (agent RemoteAgent, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &agent); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = agent.Sanitize()
	return
}

func UnmarshallLocalAgent(file []byte) (agent LocalAgent, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &agent); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	err = agent.Sanitize()
	return
}

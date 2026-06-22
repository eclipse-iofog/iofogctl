package resource

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

const (
	email    = "user@domain.com"
	password = "as901yh3rinsd"
)

func TestKubernetesControlPlaneYAMLNoEcnViewer(t *testing.T) {
	trustProxy := true
	cp := KubernetesControlPlane{
		Endpoint:   "https://controller.example.com",
		KubeConfig: "/tmp/kubeconfig",
		IofogUser:  IofogUser{Email: email},
		Controller: ControllerConfig{
			PublicUrl:  "https://controller.example.com",
			ConsoleUrl: "https://controller.example.com",
			TrustProxy: &trustProxy,
		},
	}
	out, err := yaml.Marshal(cp)
	if err != nil {
		t.Fatal(err)
	}
	text := string(out)
	for _, retired := range []string{"ecnViewerPort", "ecnViewerUrl"} {
		if strings.Contains(text, retired) {
			t.Fatalf("YAML must not contain %q:\n%s", retired, text)
		}
	}
	for _, want := range []string{"endpoint:", "publicUrl:", "consoleUrl:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("YAML missing %q:\n%s", want, text)
		}
	}
}

func TestKubernetesControlPlane(t *testing.T) {
	trustProxy := true
	cp := KubernetesControlPlane{
		Endpoint:   "123.123.123.123",
		KubeConfig: "~/.kube/config",
		IofogUser:  IofogUser{Email: "user@domain.com", Password: "password"},
		Controller: ControllerConfig{
			PublicUrl:  "https://controller.example.com",
			ConsoleUrl: "https://console.example.com",
			TrustProxy: &trustProxy,
		},
		Auth: Auth{
			Mode: "embedded",
			Bootstrap: &AuthBootstrap{
				Username: "admin",
				Password: "BootstrapPass1!",
			},
		},
	}
	if cp.Controller.PublicUrl != "https://controller.example.com" {
		t.Error("Wrong publicUrl")
	}
	if cp.Auth.Mode != "embedded" {
		t.Error("Wrong auth mode")
	}
	if endpoint, err := cp.GetEndpoint(); err != nil || endpoint != "123.123.123.123" {
		t.Error("Wrong endpoint")
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if err := cp.AddController(&KubernetesController{
		PodName:  "pod1",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}
	if err := cp.AddController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}

	if err := cp.AddController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err == nil {
		t.Error("Should have failed adding duplicate Controller")
	}

	if err := cp.UpdateController(&KubernetesController{
		PodName:  "pod2",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}
	for _, ctrl := range cp.GetControllers() {
		if ctrl.GetCreatedTime() != "now" || ctrl.GetEndpoint() != "123.123.123.123" {
			t.Error("Controller details are wrong")
		}
	}
	if err := cp.DeleteController("pod2"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}
	if err := cp.DeleteController("pod2"); err == nil {
		t.Error("Deleted non existent Controller")
	}
	if err := cp.DeleteController("pod1"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
}

func TestRemoteControlPlane(t *testing.T) {
	cp := RemoteControlPlane{
		IofogUser: IofogUser{Email: "user@domain.com", Password: "password"},
	}
	if err := cp.AddController(&RemoteController{
		Name:     "ctrl1",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}
	if endpoint, err := cp.GetEndpoint(); err != nil || endpoint != "123.123.123.123" {
		t.Error("Wrong endpoint")
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if err := cp.AddController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}

	if err := cp.AddController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "223.223.223.223",
		Created:  "now",
	}); err == nil {
		t.Error("Should have failed adding duplicate Controller")
	}

	if err := cp.UpdateController(&RemoteController{
		Name:     "ctrl2",
		Endpoint: "123.123.123.123",
		Created:  "now",
	}); err != nil {
		t.Error(err)
	}

	if len(cp.GetControllers()) != 2 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 2)
	}
	for _, ctrl := range cp.GetControllers() {
		if ctrl.GetCreatedTime() != "now" || ctrl.GetEndpoint() != "123.123.123.123" {
			t.Error("Controller details are wrong")
		}
	}
	if err := cp.DeleteController("ctrl2"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}
	if err := cp.DeleteController("ctrl2"); err == nil {
		t.Error("Deleted non existent Controller")
	}
	if err := cp.DeleteController("ctrl1"); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
}

func TestLocalControlPlane(t *testing.T) {
	cp := LocalControlPlane{
		IofogUser: IofogUser{Email: "user@domain.com", Password: "password"},
	}
	if err := cp.AddController(&LocalController{
		Name:    "ctrl1",
		Created: "now",
	}); err != nil {
		t.Error(err)
	}
	_ = cp.Sanitize()

	if endpoint, err := cp.GetEndpoint(); err != nil || !util.IsLocalHost(endpoint) {
		t.Errorf("Wrong endpoint: %s", endpoint)
	}
	if user := cp.GetUser(); user.Email != "user@domain.com" || user.Password != "password" {
		t.Error("Wrong user details")
	}
	if err := cp.Sanitize(); err != nil {
		t.Error("Failed to sanitize")
	}
	if len(cp.GetControllers()) != 1 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 1)
	}

	if err := cp.DeleteController(""); err != nil {
		t.Error(err)
	}
	if len(cp.GetControllers()) != 0 {
		t.Errorf("Controller count is wrong, %d vs %d", len(cp.GetControllers()), 0)
	}
	if ctrl, err := cp.GetController(""); err == nil || ctrl != nil {
		t.Error("Should have returned error when getting Local Controller")
	}
}

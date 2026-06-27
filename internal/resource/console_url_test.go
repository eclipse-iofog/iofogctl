package resource

import (
	"testing"
)

func TestResolveConsoleURLExplicitConsole(t *testing.T) {
	cp := &LocalControlPlane{
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{
				ConsoleUrl: "http://192.168.1.6",
				PublicUrl:  "http://192.168.1.6:51121",
			},
		},
		Endpoint: "http://192.168.1.6:51121",
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://192.168.1.6" {
		t.Fatalf("got %q, want %q", got, "http://192.168.1.6")
	}
}

func TestResolveConsoleURLRemoteFixture(t *testing.T) {
	cp := &RemoteControlPlane{
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{
				ConsoleUrl: "https://192.168.105.2:80",
				PublicUrl:  "https://192.168.105.2:51121",
			},
		},
		Endpoint: "https://192.168.105.2:51121",
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://192.168.105.2:80" {
		t.Fatalf("got %q, want %q", got, "https://192.168.105.2:80")
	}
}

func TestResolveConsoleURLHTTPAPIFallback(t *testing.T) {
	cp := &LocalControlPlane{
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{
				PublicUrl: "http://192.168.1.6:51121",
			},
		},
		Endpoint: "http://192.168.1.6:51121",
		Controllers: []LocalController{{
			Name:     "iofog",
			Endpoint: "http://192.168.1.6:51121",
		}},
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://192.168.1.6" {
		t.Fatalf("got %q, want %q", got, "http://192.168.1.6")
	}
}

func TestResolveConsoleURLHTTPSAPIFallback(t *testing.T) {
	https := true
	cp := &RemoteControlPlane{
		TLS: &ControlPlaneTLS{Cert: "cert", Key: "key"},
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{
				PublicUrl: "https://10.0.0.5:51121",
			},
		},
		Endpoint: "https://10.0.0.5:51121",
		Controllers: []RemoteController{{
			Name:     "controlplane",
			Endpoint: "https://10.0.0.5:51121",
		}},
	}
	cp.Controller.Https = &https

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://10.0.0.5:80" {
		t.Fatalf("got %q, want %q", got, "https://10.0.0.5:80")
	}
}

func TestResolveConsoleURLLocalhostAPIFallback(t *testing.T) {
	cp := &LocalControlPlane{
		Endpoint: "http://localhost:51121",
		Controllers: []LocalController{{
			Name:     "local",
			Endpoint: "http://localhost:51121",
		}},
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://localhost" {
		t.Fatalf("got %q, want %q", got, "http://localhost")
	}
}

func TestResolveConsoleURLK8sIngressHostname(t *testing.T) {
	cp := &KubernetesControlPlane{
		Controller: ControllerConfig{
			PublicUrl:  "https://controller.example.com",
			ConsoleUrl: "https://console.example.com",
		},
		Endpoint: "https://controller.example.com",
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://console.example.com" {
		t.Fatalf("got %q, want %q", got, "https://console.example.com")
	}
}

func TestResolveConsoleURLK8sFallbackWithoutPort(t *testing.T) {
	cp := &KubernetesControlPlane{
		Controller: ControllerConfig{
			PublicUrl: "https://controller.example.com",
		},
		Endpoint: "https://controller.example.com",
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://controller.example.com" {
		t.Fatalf("got %q, want %q", got, "https://controller.example.com")
	}
}

func TestResolveConsoleURLConnectOnlyEndpoint(t *testing.T) {
	cp := &RemoteControlPlane{
		Endpoint: "http://203.0.113.10:51121",
		Controllers: []RemoteController{{
			Name:     "remote",
			Endpoint: "http://203.0.113.10:51121",
		}},
	}

	got, err := ResolveConsoleURL(cp)
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://203.0.113.10" {
		t.Fatalf("got %q, want %q", got, "http://203.0.113.10")
	}
}

func TestBackfillConsoleURL(t *testing.T) {
	cp := &LocalControlPlane{
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{
				PublicUrl: "http://192.168.1.6:51121",
			},
		},
		Endpoint: "http://192.168.1.6:51121",
	}

	if err := BackfillConsoleURL(cp); err != nil {
		t.Fatal(err)
	}
	if cp.Controller.ConsoleUrl != "http://192.168.1.6" {
		t.Fatalf("ConsoleUrl = %q, want %q", cp.Controller.ConsoleUrl, "http://192.168.1.6")
	}

	original := cp.Controller.ConsoleUrl
	if err := BackfillConsoleURL(cp); err != nil {
		t.Fatal(err)
	}
	if cp.Controller.ConsoleUrl != original {
		t.Fatalf("BackfillConsoleURL overwrote explicit value: %q", cp.Controller.ConsoleUrl)
	}
}

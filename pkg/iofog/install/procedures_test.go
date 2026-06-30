package install

import "testing"

func TestEntrypointGetCommandQuotesInstallArgs(t *testing.T) {
	ep := Entrypoint{
		destPath: "/tmp/edgelet-scripts/install.sh",
		Args:     []string{"--version=v1.0.0-rc.8", "--arch=arm64", "--skip-start"},
	}
	got := ep.getCommand()
	want := "/tmp/edgelet-scripts/install.sh '--version=v1.0.0-rc.8' '--arch=arm64' '--skip-start'"
	if got != want {
		t.Fatalf("getCommand() = %q, want %q", got, want)
	}
}

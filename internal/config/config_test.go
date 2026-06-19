package config

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func TestLatestAPIVersionFromLdflag(t *testing.T) {
	if LatestAPIVersion != util.GetCliApiVersion() {
		t.Fatalf("LatestAPIVersion = %q, want %q from ldflag", LatestAPIVersion, util.GetCliApiVersion())
	}
}

func TestConfigPathUsesIofogV3(t *testing.T) {
	const want = ".iofog/v3"
	if defaultDirname != want {
		t.Fatalf("defaultDirname = %q, want %q", defaultDirname, want)
	}
}

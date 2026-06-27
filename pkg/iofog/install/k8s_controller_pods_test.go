package install

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestControllerPodLabelsMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		labels         map[string]string
		componentLabel string
		wantMatch      bool
	}{
		{
			name:           "datasance operator controller pod",
			componentLabel: "datasance.com/component",
			labels: map[string]string{
				"datasance.com/component":     "controller",
				"app.kubernetes.io/component": "controller",
			},
			wantMatch: true,
		},
		{
			name:           "iofog eclipse controller pod",
			componentLabel: "iofog.org/component",
			labels: map[string]string{
				"iofog.org/component": "controller",
			},
			wantMatch: true,
		},
		{
			name:           "wrong component value",
			componentLabel: "datasance.com/component",
			labels: map[string]string{
				"datasance.com/component": "router",
			},
		},
		{
			name:           "legacy key only",
			componentLabel: "datasance.com/component",
			labels: map[string]string{
				"iofog.org/component": "controller",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantMatch, controllerPodLabelsMatch(tt.labels, tt.componentLabel))
		})
	}
}

package install

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestIsLiveControllerPod(t *testing.T) {
	t.Parallel()

	now := metav1.Now()
	tests := []struct {
		name string
		pod  corev1.Pod
		want bool
	}{
		{
			name: "running",
			pod:  corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}},
			want: true,
		},
		{
			name: "pending",
			pod:  corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodPending}},
			want: true,
		},
		{
			name: "failed",
			pod:  corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodFailed}},
		},
		{
			name: "succeeded",
			pod:  corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodSucceeded}},
		},
		{
			name: "unknown",
			pod:  corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodUnknown}},
		},
		{
			name: "empty phase",
			pod:  corev1.Pod{},
		},
		{
			name: "running but terminating",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			},
		},
		{
			name: "pending but terminating",
			pod: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{DeletionTimestamp: &now},
				Status:     corev1.PodStatus{Phase: corev1.PodPending},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isLiveControllerPod(tt.pod))
		})
	}
}

func TestSelectControllerPodsFiltersFailedAndTerminating(t *testing.T) {
	t.Parallel()

	now := metav1.Now()
	componentLabel := "datasance.com/component"
	controllerLabels := map[string]string{componentLabel: "controller"}
	routerLabels := map[string]string{componentLabel: "router"}

	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "controller-running", Labels: controllerLabels},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "controller-pending", Labels: controllerLabels},
			Status:     corev1.PodStatus{Phase: corev1.PodPending},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "controller-failed", Labels: controllerLabels},
			Status:     corev1.PodStatus{Phase: corev1.PodFailed},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "controller-succeeded", Labels: controllerLabels},
			Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "controller-deleting", Labels: controllerLabels, DeletionTimestamp: &now},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "router-running", Labels: routerLabels},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
	}

	selected := selectControllerPods(pods, componentLabel)
	require.Equal(t, []Pod{
		{Name: "controller-running", Status: string(corev1.PodRunning)},
		{Name: "controller-pending", Status: string(corev1.PodPending)},
	}, selected)
}

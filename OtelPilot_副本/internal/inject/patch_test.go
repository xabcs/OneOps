package inject

import (
	"encoding/json"
	"strings"
	"testing"

	evanpatch "github.com/evanphx/json-patch/v5"
	jsonpatch "gomodules.xyz/jsonpatch/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"otelpilot/internal/config"
)

func testPod(labels map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-java-7d9c8b6f5-x2k4l",
			Namespace: "default",
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "app", Image: "eclipse-temurin:17"}},
		},
	}
}

func applyOps(t *testing.T, pod *corev1.Pod, ops []jsonpatch.JsonPatchOperation) *corev1.Pod {
	t.Helper()
	raw, err := json.Marshal(pod)
	if err != nil {
		t.Fatalf("marshal pod: %v", err)
	}
	for _, op := range ops {
		b, _ := json.Marshal(op)
		patched, err := evanpatch.DecodePatch([]byte("[" + string(b) + "]"))
		if err != nil {
			t.Fatalf("decode patch %s: %v", b, err)
		}
		raw, err = patched.Apply(raw)
		if err != nil {
			t.Fatalf("apply patch %s: %v", b, err)
		}
	}
	out := &corev1.Pod{}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshal patched pod: %v", err)
	}
	return out
}

func TestShouldInject(t *testing.T) {
	cases := []struct {
		name  string
		label string
		want  bool
	}{
		{"on enables injection", EnableOn, true},
		{"off skips", "off", false},
		{"empty skips", "", false},
		{"boolean-like string skips", "true", false},
	}
	for _, c := range cases {
		pod := testPod(map[string]string{LabelEnable: c.label})
		if got := ShouldInject(pod); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
	// 无标签跳过
	if ShouldInject(testPod(nil)) {
		t.Error("pod without label must not inject")
	}
	// 防重注入
	pod := testPod(map[string]string{LabelEnable: EnableOn})
	pod.Annotations = map[string]string{AnnotationStatus: StatusInjected}
	if ShouldInject(pod) {
		t.Error("already injected pod must not re-inject")
	}
}

func TestBuildPatches(t *testing.T) {
	pod := testPod(map[string]string{
		LabelEnable:       EnableOn,
		LabelAppName:      "demo-java",
		LabelJDKVersion:   "17",
		LabelAgentVersion: "2.31.1",
	})
	ops, err := BuildPatches(pod, config.DefaultSnapshot())
	if err != nil {
		t.Fatalf("build patches: %v", err)
	}
	got := applyOps(t, pod, ops)

	// 防重注入注解
	if got.Annotations[AnnotationStatus] != StatusInjected {
		t.Errorf("missing injected annotation: %v", got.Annotations)
	}
	if got.Annotations[AnnotationAgentVersion] != "2.31.1" {
		t.Errorf("agent version annotation = %q", got.Annotations[AnnotationAgentVersion])
	}
	if got.Annotations[AnnotationJDKVersion] != "17" {
		t.Errorf("jdk version annotation = %q", got.Annotations[AnnotationJDKVersion])
	}
	// emptyDir 卷 + Init 容器
	if !hasVolume(got.Spec.Volumes, VolumeName) || got.Spec.Volumes[0].EmptyDir == nil {
		t.Error("emptyDir volume not injected")
	}
	if len(got.Spec.InitContainers) != 1 || got.Spec.InitContainers[0].Name != InitContainerName {
		t.Fatalf("init container not injected: %+v", got.Spec.InitContainers)
	}
	if !strings.Contains(strings.Join(got.Spec.InitContainers[0].Args, " "), "v2.31.1/opentelemetry-javaagent.jar") {
		t.Errorf("init container download url missing version: %v", got.Spec.InitContainers[0].Args)
	}
	// 业务容器：挂载 + 环境变量
	c := got.Spec.Containers[0]
	if len(c.VolumeMounts) != 1 || c.VolumeMounts[0].Name != VolumeName || !c.VolumeMounts[0].ReadOnly {
		t.Errorf("volume mount wrong: %+v", c.VolumeMounts)
	}
	env := map[string]corev1.EnvVar{}
	for _, e := range c.Env {
		env[e.Name] = e
	}
	if v := env[JavaToolOptions]; v.Value != "-javaagent:/otel-pilot/opentelemetry-javaagent.jar" {
		t.Errorf("JAVA_TOOL_OPTIONS = %q", v.Value)
	}
	if v := env["OTEL_SERVICE_NAME"]; v.Value != "demo-java" {
		t.Errorf("OTEL_SERVICE_NAME = %q", v.Value)
	}
	if v := env["OTEL_RESOURCE_ATTRIBUTES"]; !strings.Contains(v.Value, "k8s.pod.name=$(OTEL_POD_NAME)") {
		t.Errorf("OTEL_RESOURCE_ATTRIBUTES = %q", v.Value)
	}
	// $(VAR) 引用变量必须先于其注入
	podVarIdx, refIdx := -1, -1
	for i, e := range c.Env {
		if e.Name == "OTEL_POD_NAME" {
			podVarIdx = i
		}
		if e.Name == "OTEL_RESOURCE_ATTRIBUTES" {
			refIdx = i
		}
	}
	if podVarIdx == -1 || refIdx == -1 || podVarIdx > refIdx {
		t.Errorf("env order invalid: OTEL_POD_NAME=%d OTEL_RESOURCE_ATTRIBUTES=%d", podVarIdx, refIdx)
	}
}

func TestBuildPatchesMergesExistingJavaToolOptions(t *testing.T) {
	pod := testPod(map[string]string{LabelEnable: EnableOn})
	pod.Spec.Containers[0].Env = []corev1.EnvVar{
		{Name: JavaToolOptions, Value: "-Xmx512m"},
		{Name: "OTEL_TRACES_EXPORTER", Value: "none"}, // 业务自定义不覆盖
	}
	ops, _ := BuildPatches(pod, config.DefaultSnapshot())
	got := applyOps(t, pod, ops)

	env := map[string]corev1.EnvVar{}
	for _, e := range got.Spec.Containers[0].Env {
		env[e.Name] = e
	}
	if v := env[JavaToolOptions]; v.Value != "-Xmx512m -javaagent:/otel-pilot/opentelemetry-javaagent.jar" {
		t.Errorf("merged JAVA_TOOL_OPTIONS = %q", v.Value)
	}
	if v := env["OTEL_TRACES_EXPORTER"]; v.Value != "none" {
		t.Errorf("business OTEL_TRACES_EXPORTER overridden: %q", v.Value)
	}
}

func TestBuildPatchesExistingVolumesAndInitContainers(t *testing.T) {
	pod := testPod(map[string]string{LabelEnable: EnableOn})
	pod.Spec.Volumes = []corev1.Volume{{Name: "cache", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}}
	pod.Spec.InitContainers = []corev1.Container{{Name: "warmup", Image: "busybox"}}
	ops, _ := BuildPatches(pod, config.DefaultSnapshot())
	got := applyOps(t, pod, ops)

	if len(got.Spec.Volumes) != 2 || got.Spec.Volumes[1].Name != VolumeName {
		t.Errorf("volume append failed: %+v", got.Spec.Volumes)
	}
	if len(got.Spec.InitContainers) != 2 || got.Spec.InitContainers[1].Name != InitContainerName {
		t.Errorf("init container append failed: %+v", got.Spec.InitContainers)
	}
}

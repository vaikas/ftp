// +build e2e

package e2e

import (
	"context"
	"os"
	"testing"

	. "github.com/cloudevents/sdk-go/v2/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	sourcesv1beta1 "knative.dev/eventing/pkg/apis/sources/v1beta1"
	reconcilertestingv1beta1 "knative.dev/eventing/pkg/reconciler/testing/v1beta1"
	"knative.dev/eventing/pkg/utils"
	"knative.dev/eventing/test/lib/recordevents"
	"knative.dev/eventing/test/lib/resources"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	pkgTest "knative.dev/pkg/test"
)

func TestFTPSource(t *testing.T) {
	const (
		ftpSourceName        = "e2e-container-source"
		templateName         = "e2e-container-source-template"
		recordEventPodName   = "e2e-ftp-source-logger-pod"
		imageName            = "ftpsource"
		cloudEventsEventType = "org.aikas.ftp.fileadded"
	)
	cloudEventsSourceName := "//" + os.Getenv("FTP_URL") + "/incoming"

	matcherGen := func(cloudEventsSourceName, cloudEventsEventType string) EventMatcher {
		return AllOf(
			HasSource(cloudEventsSourceName),
			HasType(cloudEventsEventType),
		)
	}

	client := setup(t, true)
	defer tearDown(client)

	ctx := context.Background()

	if _, err := utils.CopySecret(client.Kube.CoreV1(), "default", "sftp-secret", client.Namespace, "default"); err != nil {
		t.Fatalf("could not copy secret(%s): %v", "sftp-secret", err)
	}

	if err := createCMEditRoleAndBinding(ctx, client.Kube.RbacV1(), client.Namespace); err != nil {
		t.Fatalf("could not create cm edit role and binding %v", err)
	}

	// create event record pod
	eventTracker, _ := recordevents.StartEventRecordOrFail(ctx, client, recordEventPodName)
	// create container source
	// args are the arguments passing to the container, msg is used in the heartbeats image
	args := []string{
		"--sftpServer=" + os.Getenv("FTP_URL"),
		"--secure=true",
		"--dir=/incoming",
		"--storename=sftp-store",
		"--probeFrequency=5",
	}
	// envVars are the environment variables of the container
	envVars := []corev1.EnvVar{{
		Name:  "POD_NAME",
		Value: templateName,
	}, {
		Name:  "POD_NAMESPACE",
		Value: client.Namespace,
	}, {
		Name:  "SYSTEM_NAMESPACE",
		Value: client.Namespace,
	}, {
		Name: "FTP_USER",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "user",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "sftp-secret",
				},
			},
		},
	}, {
		Name: "FTP_PASSWORD",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				Key: "password",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "sftp-secret",
				},
			},
		},
	}}
	ftpSource := reconcilertestingv1beta1.NewContainerSource(
		ftpSourceName,
		client.Namespace,
		reconcilertestingv1beta1.WithContainerSourceSpec(sourcesv1beta1.ContainerSourceSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: templateName,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            imageName,
						Image:           pkgTest.ImagePath(imageName),
						ImagePullPolicy: corev1.PullIfNotPresent,
						Args:            args,
						Env:             envVars,
					}},
				},
			},
			SourceSpec: duckv1.SourceSpec{
				Sink: duckv1.Destination{Ref: resources.KnativeRefForService(recordEventPodName, client.Namespace)},
			},
		}),
	)
	client.CreateContainerSourceV1Beta1OrFail(ftpSource)

	// wait for all test resources to be ready
	client.WaitForAllTestResourcesReadyOrFail(ctx)

	postMessage("/incoming/test.txt", client)

	eventTracker.AssertExact(1, recordevents.MatchEvent(matcherGen(cloudEventsSourceName, cloudEventsEventType)))
}

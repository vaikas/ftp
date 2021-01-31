package e2e

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	testlib "knative.dev/eventing/test/lib"
	pkgtest "knative.dev/pkg/test"
)

func postMessage(path string, client *testlib.Client) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sftp-publisher" + uuid.New().String(),
			Namespace: client.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "sftp-publisher",
						Image:           pkgtest.ImagePath("sftp-publisher"),
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{{
							Name:  "PATH",
							Value: path,
						}, {
							Name:  "FTP_URL",
							Value: os.Getenv("FTP_URL"),
						}, {
							Name:  "USER",
							Value: os.Getenv("USER"),
						}, {
							Name:  "PASS",
							Value: os.Getenv("PASS"),
						}},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	pkgtest.CleanupOnInterrupt(func() {
		client.Kube.BatchV1().Jobs(job.Namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{})
	}, client.T.Logf)
	job, err := client.Kube.BatchV1().Jobs(job.Namespace).Create(context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		client.T.Fatalf("Error creating Job: %v", err)
	}

	defer func() {
		err := client.Kube.BatchV1().Jobs(job.Namespace).Delete(context.Background(), job.Name, metav1.DeleteOptions{})
		if err != nil {
			client.T.Errorf("Error cleaning up Job %s", job.Name)
		}
	}()

	// Wait for the Job to report a successful execution.
	waitErr := wait.PollImmediate(1*time.Second, 2*time.Minute, func() (bool, error) {
		js, err := client.Kube.BatchV1().Jobs(job.Namespace).Get(context.Background(), job.Name, metav1.GetOptions{})
		if apierrs.IsNotFound(err) {
			return false, nil
		} else if err != nil {
			return true, err
		}

		client.T.Logf("Active=%d, Failed=%d, Succeeded=%d", js.Status.Active, js.Status.Failed, js.Status.Succeeded)

		// Check for successful completions.
		return js.Status.Succeeded > 0, nil
	})
	if waitErr != nil {
		client.T.Fatalf("Error waiting for Job to complete successfully: %v", waitErr)
	}
}

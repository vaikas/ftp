module github.com/vaikas/ftp

go 1.15

replace (
	k8s.io/api => k8s.io/api v0.19.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.19.7
	k8s.io/client-go => k8s.io/client-go v0.19.7
)

require (
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/google/uuid v1.1.2
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/pkg/sftp v1.12.0
	github.com/secsy/goftp v0.0.0-20200609142545-aa2de14babf4
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/mod v0.4.0 // indirect
	golang.org/x/tools v0.0.0-20210113180300-f96436850f18 // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	k8s.io/api v0.19.7
	k8s.io/apimachinery v0.19.7
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210113233702-8566a335510f // indirect
	knative.dev/eventing v0.20.1
	knative.dev/pkg v0.0.0-20210107022335-51c72e24c179
	knative.dev/reconciler-test v0.0.0-20210108100436-db4d65735605
)

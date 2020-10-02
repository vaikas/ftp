module github.com/vaikas/ftp

go 1.15

replace (
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
)

require (
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/google/uuid v1.1.2
	github.com/pkg/sftp v1.12.0
	github.com/secsy/goftp v0.0.0-20200609142545-aa2de14babf4
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201001193750-eb9a90e9f9cb
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.18.0
	knative.dev/pkg v0.0.0-20201001022229-7c082ac1c01d
)

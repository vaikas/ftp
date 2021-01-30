module github.com/vaikas/ftp

go 1.15

replace (
	k8s.io/api => k8s.io/api v0.18.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.15
	k8s.io/client-go => k8s.io/client-go v0.18.15
)

require (
	github.com/Azure/azure-sdk-for-go v42.3.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.10.2 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/cloudevents/sdk-go/v2 v2.3.1
	github.com/go-training/helloworld v0.0.0-20200225145412-ba5f4379d78b // indirect
	github.com/google/go-containerregistry v0.4.0 // indirect
	github.com/google/ko v0.7.1-0.20210121183014-82cabb40bae5 // indirect
	github.com/google/uuid v1.1.2
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/pkg/sftp v1.12.0
	github.com/secsy/goftp v0.0.0-20200609142545-aa2de14babf4
	github.com/vdemeester/k8s-pkg-credentialprovider v1.18.1-0.20201019120933-f1d16962a4db // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/mod v0.4.0 // indirect
	golang.org/x/tools v0.0.0-20210113180300-f96436850f18 // indirect
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.18.15
	k8s.io/apimachinery v0.19.6
	k8s.io/cli-runtime v0.18.8 // indirect
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.20.1
	knative.dev/pkg v0.0.0-20210107022335-51c72e24c179
	sigs.k8s.io/kind v0.8.1 // indirect
	sigs.k8s.io/structured-merge-diff v1.0.1-0.20191108220359-b1b620dd3f06 // indirect
)

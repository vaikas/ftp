# Knative`Twitter Source` CRD.

## Overview

This repository implements a simple Event Source for wiring Twitter events
into [Knative Eventing](http://github.com/knative/eventing).

## Details

This uses containersource

## Prerequisites

1. Create a
   [Google Cloud project](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
   and install the `gcloud` CLI and run `gcloud auth login`. This sample will
   use a mix of `gcloud` and `kubectl` commands. The rest of the sample assumes
   that you've set the `$PROJECT_ID` environment variable to your Google Cloud
   project id, and also set your project ID as default using
   `gcloud config set project $PROJECT_ID`.

1. Setup [Knative Serving](https://github.com/knative/docs/blob/master/install)

1. Configure [outbound network access](https://github.com/knative/docs/blob/master/serving/outbound-network-access.md)

1. Setup [Knative Eventing](https://github.com/knative/docs/tree/master/eventing)
   using the `release.yaml` file. This example does not require GCP.

1. Have Twitter API access keys. [Good instructions on how to get them](https://iag.me/socialmedia/how-to-create-a-twitter-app-in-8-easy-steps/)

## Create a secret with your Twitter secrets

Modify (or create a file like this) ./secret.yaml and replace TWITTER_* entries with real entries and then create the secret:

```shell
apiVersion: v1
kind: Secret
metadata:
  name: twitter-secret
type: Opaque
stringData:
  consumer-key: TWITTER_CONSUMER_KEY
  consumer-secret-key: TWITTER_CONSUMER_SECRET_KEY
  access-token: TWITTER_ACCESS_TOKEN
  access-secret: TWITTER_ACCESS_SECRET
```


And then create the secret like so:
```shell
kubectl create -f ./secret.yaml
```

## Launch the twitter source

```shell

```

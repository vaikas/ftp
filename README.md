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

## Create the function that receives the tweets matching your query
We're going to launch a service in Knative that gets invoked for each of the tweets
matching your query term that we'll set up below.

```shell
kubectl --namespace default apply -f https://raw.githubusercontent.com/vaikas-google/twitter/master/config/service.yaml
```

## Create a secret with your Twitter secrets

Modify (or create a file like this) ./secret.yaml and replace TWITTER_* entries with real entries
for your account and then create the secret:

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
The source expects you to specify a query string that tells what to search for in the sea
of tweets. For example, if you wanted to look for `knative`, you'd do:

```shell
curl https://raw.githubusercontent.com/vaikas-google/twitter/master/config/search-source.yaml | \
sed "s/QUERY/knative/g" | kubectl apply -f -
```

if you want to search for something else, replace knative with the query string you want
to look for.
```shell
curl https://raw.githubusercontent.com/vaikas-google/twitter/master/config/search-source.yaml | \
sed "s/QUERY/yourquerystring/g" | kubectl apply -f -
```

## Look for the results of your function execution

```shell
kubectl -l 'serving.knative.dev/service=twitter-dumper' logs -c user-container
```

and you should see tweets that match your query string. When I look for knative, I might see things like this:

```shell
2019/01/11 23:03:10 Received Cloud Event Context as: {CloudEventsVersion:0.1 EventID:1083397354315792386 EventTime:2019-01-10 16:17:12 +0000 UTC EventType:com.twitter EventTypeVersion: SchemaURL: ContentType:application/json Source:com.twitter Extensions:map[]}
2019/01/11 23:03:10 Got tweet from "Serverless Fan" text: "RT @sarbjeetjohal: Lambda comes to your data enter through #knative lambda runtime! We knew it was matter of days for people to spread it l…"
2019/01/11 23:03:10 Received Cloud Event Context as: {CloudEventsVersion:0.1 EventID:1083397331918106625 EventTime:2019-01-10 16:17:07 +0000 UTC EventType:com.twitter EventTypeVersion: SchemaURL: ContentType:application/json Source:com.twitter Extensions:map[]}
2019/01/11 23:03:10 Got tweet from "Sarbjeet Johal" text: "Lambda comes to your data enter through #knative lambda runtime! We knew it was matter of days for people to spread… https://t.co/aWo8BEh7GS"
2019/01/11 23:03:10 Received Cloud Event Context as: {CloudEventsVersion:0.1 EventID:1083393535674601472 EventTime:2019-01-10 16:02:02 +0000 UTC EventType:com.twitter EventTypeVersion: SchemaURL: ContentType:application/json Source:com.twitter Extensions:map[]}
2019/01/11 23:03:10 Got tweet from "David Metcalfe" text: "RT @RedHatPartners: #RedHat collaborates with @Google, @SAP, @IBM and others on @KnativeProject to deliver hybrid #serverless workloads to…"
```

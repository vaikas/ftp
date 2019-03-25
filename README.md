# Knative`FTP/SFTP Source` CRD.

## Overview

This repository implements a simple Event Source for looking at
an FTP / SFTP server and creating notifications for new files
uploaded into [Knative Eventing](http://github.com/knative/eventing).

## Details

This uses containersource. Until Istio 1.1 you'll need to annotate the
FTP source with the `traffic.sidecar.istio.io/includeOutboundIPRanges`
annotation to ensure the source can emit events into the mesh.

## Prerequisites

1. Setup [Knative Serving](https://github.com/knative/docs/blob/master/docs/install)
  This is used for the function that gets subscribed to the new files notifications.

1. Configure [outbound network access](https://github.com/knative/docs/blob/master/docs/serving/outbound-network-access.md)
  **note** you will need to determine the IP range of your cluster. So determine the IP range of your cluster. For example
  if your IP ranges are: `10.16.0.0/14,10.19.240.0/20` export them like so.
```shell
export INCLUDE_OUTBOUND_IPRANGES="10.16.0.0/14,10.19.240.0/20"
```
  
1. Setup [Knative Eventing](https://github.com/knative/docs/tree/master/docs/eventing)
   using the `release.yaml` file. This example does not require GCP.

1. Have user/password for the FTP/SFTP server. Only user/password is supported now (PRs accepted ;) )

1. Permit the service account the source runs as to read/modify ConfigMaps. This is necessary as the
   source uses ConfigMaps to store it's state. If you do **not** run as the normal service account default/default
   as per the instructions below, you will need to modify the following file to modify permissions appropriately. By
   default the file below will grant default service account in the default namespace rights to Read/Write Configmaps.
   
```shell
kubectl --namespace default apply -f https://raw.githubusercontent.com/vaikas-google/ftp/master/config/cm_role.yaml
```

## Create the function that receives the notifications about new files being uploaded
We're going to launch a service in Knative that gets invoked for each of the new
files being uploaded

```shell
kubectl --namespace default apply -f https://raw.githubusercontent.com/vaikas-google/ftp/master/config/service.yaml
```

## Create a secret with your FTP credentials OR with your SFTP credentials

### FTP Credentials

Modify (or create a file like this) ./config/ftp-secret.yaml and replace FTP_* entries with real entries
for your account and then create the secret:

```shell
apiVersion: v1
kind: Secret
metadata:
  name: ftp-secret
type: Opaque
stringData:
  user: FTP_USER
  password: FTP_PASSWORD
```


And then create the secret like so:
```shell
kubectl create -f ./secret.yaml
```

Or you can do the following using checked in files (assuming your username is `myusername` and password is `mypassword`:
```shell
curl https://raw.githubusercontent.com/vaikas-google/ftp/master/config/ftp-secret.yaml | \
sed "s/FTP_USER/myusername/g" | \
sed "s/FTP_PASSWORD/mypassword/g" | \
kubectl apply -f -
```


### SFTP Credentials

Modify (or create a file like this) ./config/sftp-secret.yaml and replace SFTP_* entries with real entries
for your account and then create the secret:

```shell
apiVersion: v1
kind: Secret
metadata:
  name: sftp-secret
type: Opaque
stringData:
  user: SFTP_USER
  password: SFTP_PASSWORD
```


And then create the secret like so:
```shell
kubectl create -f ./secret.yaml
```

Or you can do the following using checked in files (assuming your username is `myusername` and password is `mypassword`:
```shell
curl https://raw.githubusercontent.com/vaikas-google/ftp/master/config/sftp-secret.yaml | \
sed "s/SFTP_USER/myusername/g" | \
sed "s/SFTP_PASSWORD/mypassword/g" | \
kubectl apply -f -
```

## Launch the FTP / SFTP source
The source needs to communicate to outside FTP / SFTP server and due to Istio side car injection
it's unable to do unless you specify which the internal networks are (then outbound connection is good).
So, as stated above, make sure you've set `INCLUDE_OUTBOUND_IPRANGES` env variable to your internal IP range. 
you also need to specify which directory to watch and of course which server to connect to.
I've used this for testing myself:
server: `ftp1.at.proftpd.org`
dir: /devel/source

**NOTE** for testing with the above server, I used these credentials for my secret.
user: `anonymous`
password: `myemailhere.example.com`

So, with those settings, for **FTP**, you'd run:

```shell
curl https://raw.githubusercontent.com/vaikas-google/ftp/master/config/ftp-watcher-source.yaml | \
sed "s/INCLUDE_OUTBOUND_IPRANGES/$INCLUDE_OUTBOUND_IPRANGES/g" | \
sed "s/FTP_SERVER/ftp1.at.proftpd.org/g" | \
sed "s/fTP_DIR/"/devel/source"/g" | \
kubectl apply -f -
```

For **SFTP** you have to use the sftp-watcher-source.yaml, so same as before, youd do:
server: `test.rebex.net
dir: /pub/example

**NOTE** for testing with the above server, I used these credentials for my secret.
user: `demo`
password: password

```shell
curl https://raw.githubusercontent.com/vaikas-google/ftp/master/config/sftp-watcher-source.yaml | \
sed "s/INCLUDE_OUTBOUND_IPRANGES/$INCLUDE_OUTBOUND_IPRANGES/g" | \
sed "s/SFTP_SERVER/test.rebex.net/g" | \
sed "s/fTP_DIR/"/pub/example"/g" | \
kubectl apply -f -
```

## Look for the results of your function execution

You might have to wait some seconds while the elves are busily fetching your tweets, be patient...

```shell
kubectl -l 'serving.knative.dev/service=ftp-dumper' logs -c user-container
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

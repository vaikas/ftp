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
sed "s@INCLUDE_OUTBOUND_IPRANGES@$INCLUDE_OUTBOUND_IPRANGES@g" | \
sed "s@FTP_SERVER@ftp1.at.proftpd.org@g" | \
sed "s@FTP_DIR@devel/source@g" | \
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
sed "s@INCLUDE_OUTBOUND_IPRANGES@$INCLUDE_OUTBOUND_IPRANGES@g" | \
sed "s@SFTP_SERVER@test.rebex.net:22@g" | \
sed "s@SFTP_DIR@/pub/example@g" | \
kubectl apply -f -
```

## Look for the results of your function execution

You might have to wait some seconds while the elves are busily fetching your tweets, be patient...

```shell
kubectl -l 'serving.knative.dev/service=ftp-dumper' logs -c user-container
```

and you should see tweets that match your query string. When I look for knative, I might see things like this:

```shell
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190323.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:3bf88bb5-db1a-4955-bb38-dbe23a70531b Time:2019-03-23T00:07:57Z SchemaURL: ContentType:0xc0002260a0 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190323.tar.gz Size:7571337 ModTime:2019-03-23 00:07:57 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190321.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:c2159d70-a5df-4769-a3f9-0c77b3492fea Time:2019-03-21T00:10:55Z SchemaURL: ContentType:0xc0000d5da0 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190321.tar.gz Size:7571222 ModTime:2019-03-21 00:10:55 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190318.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:d152850a-95b8-4856-a2a6-9bec8ede0a4a Time:2019-03-18T00:07:33Z SchemaURL: ContentType:0xc0000100a0 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190318.tar.gz Size:7571369 ModTime:2019-03-18 00:07:33 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190324.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:1513477f-5d7e-485a-9357-cfc9277afa55 Time:2019-03-24T00:08:11Z SchemaURL: ContentType:0xc000010180 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190324.tar.gz Size:7571258 ModTime:2019-03-24 00:08:11 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190322.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:c800a630-6972-463e-9e3c-48c393be3831 Time:2019-03-22T00:08:46Z SchemaURL: ContentType:0xc000226270 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190322.tar.gz Size:7571216 ModTime:2019-03-22 00:08:46 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190320.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:1cb6445f-45e8-4ecb-a6ad-95b9e78c95bf Time:2019-03-20T00:07:48Z SchemaURL: ContentType:0xc0000d5ea0 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190320.tar.gz Size:7571323 ModTime:2019-03-20 00:07:48 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190325.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:1fe8f294-3cef-483a-8072-fcdeedced34e Time:2019-03-25T00:08:45Z SchemaURL: ContentType:0xc0000d5f80 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190325.tar.gz Size:7571243 ModTime:2019-03-25 00:08:45 +0000 UTC}
----------------------------
Got Event Context: {SpecVersion:0.2 Type:org.aikas.ftp.fileadded Source:{URL:{Scheme:ftp Opaque: User: Host:ftp1.at.proftpd.orgdevel Path:/source/proftpd-cvs-20190319.tar.gz RawPath: ForceQuery:false RawQuery: Fragment:}} ID:29c38bb5-3fd1-474a-8205-109af5122a08 Time:2019-03-19T00:06:43Z SchemaURL: ContentType:0xc000010260 Extensions:map[]}
Got Data: {Name:proftpd-cvs-20190319.tar.gz Size:7571362 ModTime:2019-03-19 00:06:43 +0000 UTC}
```

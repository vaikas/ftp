# Knative `FTP/SFTP Source` CRD.

## Overview

This repository implements a simple Event Source for looking at
an FTP / SFTP server and creating notifications for new files
uploaded.

## Details

This uses [containersource](https://knative.dev/docs/eventing/samples/container-source/). 

## Prerequisites

1. Setup [Knative Eventing Core](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-eventing-component)
  
   CRD's and Core components should suffice.

2. Setup [Knative Serving](https://knative.dev/docs/install/any-kubernetes-cluster/#installing-the-serving-component)
  
   This step is needed only if sink is Knative service. If it is regular kubernetes service , it is not needed.

3. [ko](https://github.com/google/ko)

4. FTP Server
  
   For `testing` you could use the [simple-ftp-server](./config/200-ftp.yaml) in config.

   ```shell
    cat config/200-ftp.yaml | \
    sed "s/FTP_USER/myusername/g" | \
    sed "s/FTP_PASS/mypassword/g" | \
    kubectl apply --namespace default -f -
   ```

5. Permit the service account the source runs as to read/modify ConfigMaps. This is necessary as the
   source uses ConfigMaps to store it's state. If you do **not** run as the normal service account default/default
   as per the instructions below, you will need to modify the following file to modify permissions appropriately. By
   default the file below will grant default service account in the default namespace rights to Read/Write Configmaps.
   
   ```shell
    kubectl --namespace default apply -f config/100-cm_role.yaml
   ```

## Create the sink(knative service) that receives the notifications about new files being uploaded
    
```shell
  ko --namespace default apply -f config/300-service.yaml
```

## Create a secret with your FTP credentials OR with your SFTP credentials

### FTP Credentials

Modify (or create a file like this) ./config/300-sftp-secret.yaml and replace FTP_* entries with real entries
for your account and then create the secret:

```shell
  cat config/300-sftp-secret.yaml | \
  sed "s/FTP_USER/myusername/g" | \
  sed "s/FTP_PASS/mypassword/g" | \
  kubectl apply --namespace default -f -
```

## Launch the FTP / SFTP source
 
Please checkout the args that can be given to the FTP source in config/400-sftp-watcher-source.yaml.

```shell
  cat config/400-sftp-watcher-source.yaml | \
  sed "s@SFTP_SERVER@$(kubectl get svc my-ftp-service --namespace default -ojsonpath='{.spec.clusterIP}')@g" | \
  sed "s@SFTP_PORT@$(kubectl get svc my-ftp-service --namespace default -ojsonpath='{.spec.ports[0].port}')@g" | \
  sed "s@MONITOR_DIRECTORY@/incoming@g" | \
  ko apply -f -
```

## Look for the results of your function execution

Load files in ftp server and you will see logs similar to below in ftp-dumper

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

[![Build Status](https://travis-ci.org/emacski/k8s-app-config.svg?branch=master)](https://travis-ci.org/emacski/k8s-app-config)

k8s-app-config
--------------

A very simple application config helper CLI utility. Useful for bootstrapping applications in kubernetes where the
application needs to be aware of the runtime state of other components in kubernetes.

Note: This is considered a prototype and the CLI is likely to change

The inspiration for this utility is based on the functionality of
https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/fluentd-elasticsearch/es-image/elasticsearch_logging_discovery.go
(Copyright 2017 The Kubernetes Authors. [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0))

The idea here is that this functionality might be useful for configuring other appliations in the same manner.

**Example**

Configuring the discovery hosts for elasticsearch.

The `hosts` command allows for the retrieval of host (pod) ips of a given service with the capability to wait for a
minimum number of pods to register with the service. This means we don't have to have any prior knowledge of pod
level ips to configure an elasticsearch cluster.

```bash
$ k8s-app-config hosts --namespace kube-system --service elasticsearch-loging \
    --wait 5 --min-hosts 2 --format 'discovery.zen.ping.unicast.hosts: "{{ join .hosts "," }}"'
```

The above will wait 5 minutes for a minimum of 2 hosts (pods) to register with the specified service

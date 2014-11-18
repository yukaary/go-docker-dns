# A Simple DNS Server

For a docker container written by Go.
This DNS server is implemeted based on [miekg/dns](https://github.com/miekg/dns).

## Purpose

To work with a docker tool like a [fig](http://www.fig.sh/index.html) - isolated development environments using Docker -, this program will help a name resolving when you scale a something like a clustered service.

First of all, my motivation to make this is building a rabbitMQ cluster over a 1 coreOS machine using fig. Currently docker itelf has a limitation to link nighborhood containers. There are a way to link a container at starting with `--link` option. It can link a working container, but not working one. While using a fig, RabbitMQ try to access another nodes when a new node is added.


Here is a sample `fig.yml`.
```
rabbit:
    image: bijukunjummen/rabbitmq-server
    hostname: rabbit_1
    ports:
        - "5672:5672"
        - "15672:15672"
mqnode:
    image: bijukunjummen/rabbitmq-server
    environment:
        - CLUSTERED=true
        - CLUSTER_WITH=rabbit_1
        - RAM_NODE=true
```

Using this configuration, fig can link a 1 master and 1 slave. But multiple slaves are not available because of the failer of name resolving.

```
$ fig start rabbit
# it works
$ fig start mqnode
# it does not works as expected
$ fig scale mqnode=2
```

`fig scale mqnode=2` fails to join a second node with an error like this.
```
mqnode_2 | Clustering node rabbit@099c212478e7 with rabbit@rabbit_1 ...
mqnode_1 | 
mqnode_1 | =WARNING REPORT==== 18-Nov-2014::18:51:55 ===
mqnode_1 | global: rabbit@bc4193e30191 failed to connect to rabbit@099c212478e7
```

mqnode_1 can resolve a name of a master node, but mqnoe_1 can not see mqnode-2 because it's not exist when mqnode-1 is started.

## Approachs

To ease a limitation above, I try to implement a simple DNS function using Go and it's DNS library. When this program started , it watchs [Docker remote API - /events](https://docs.docker.com/reference/api/docker_remote_api_v1.15/) continueously. And if new container is started, then it `inspect` an information of this container and extract a hostname and it's IP address. These are stored into an in-memory, and when an another container requests a name resolving, it returns a record if it found.

new `fig.yml`
```
rabbit:
    image: bijukunjummen/rabbitmq-server
    hostname: rabbit_1
    ports:
        - "5672:5672"
        - "15672:15672"
    dns:
        - "IP addr over running this program"
mqnode:
    image: bijukunjummen/rabbitmq-server
    environment:
        - CLUSTERED=true
        - CLUSTER_WITH=rabbit_1
        - RAM_NODE=true
    dns:
        - "IP addr over running this program"
```

Start this program.
(it's just a prototype. So please install depended libraries by `go get`)

```
$ sudo su
$ export GOROOT=/usr/local/go
$ export GOPATH=/home/xxxx/go
$ go run main.go -stderrthreshold=INFO -url $DOCKER_HOST
```

* When DNSMasq is activate, it will fail to bind a port 53.

After that, start rabbitMQ cluster.
```
$ fig start rabbit
$ fig start mqnode
$ fig scale mqnode=2
```

Then access to `http://your.docker.host:15672` will contain a 1 master node and 2 clusters.

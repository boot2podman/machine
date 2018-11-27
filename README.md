# Podman Machine

Machine lets you create Podman hosts on your computer.
It creates servers with Podman on them, then
configures the Podman client to talk to them.

## Getting Started

``` console
$ podman-machine create -d virtualbox box
Running pre-create checks...
Creating machine...
(box) Creating VirtualBox VM...
(box) Creating SSH key...
(box) Starting the VM...
(box) Check network to re-create if needed...
(box) Waiting for an IP...
Machine "box" was started.
```

## Connecting


### podman

You can run the `podman` command over ssh:

``` console
$ podman-machine ssh box -- sudo podman version
```

### pypodman

Or you can use the `pypodman` tool remotely:

``` bash
$ eval $(podman-machine env box)
$ pypodman version
$ pypodman --help
```

This will use environment variables to connect.

See https://github.com/containers/libpod/tree/master/contrib/python

### varlink

Connect directly with `varlink` over the bridge:

``` bash
$ eval $(podman-machine env box --varlink)
$ varlink call io.podman.GetVersion
$ varlink help io.podman
```

You might need `--bridge="$VARLINK_BRIDGE"`.

See https://github.com/varlink/libvarlink/tree/master/tool

## Driver Plugins

These core driver plugins are bundled:

* None
* VirtualBox
* QEMU (KVM)

It is possible to add standalone drivers.

## Cloud Drivers

Cloud drivers are explicitly **not** supported.

Please use [Kubernetes](https://kubernetes.io) for that, instead.

## Inspiration

Podman Machine is inspired by [Docker Machine](https://github.com/docker/machine), which is
a similar solution but for another popular container runtime.

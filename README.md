# Podman Machine

Machine lets you create Podman hosts on your computer.
It creates servers with Podman on them, then
configures the Podman client to talk to them.

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

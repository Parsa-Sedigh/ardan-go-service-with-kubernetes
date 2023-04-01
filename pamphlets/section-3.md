## 10-3.0 Intro
## 11-3.1: Tooling Installation
We're gonna use `Kind` for local k8s env.

Kustomize comes with the installation of kubectl. But we want to install kubectl separately because we're gonna use it separately.

You can install these with brew.

## 12-3.2: Understanding Clusters, Nodes and Pods
k8s is about abstracting away from compute power, machines, racks and OS. This is what cloud is really all about, abstracting hardware.

Think of a cluster as your compute power. Everything's gonna run inside this cluster. 

Node is a machine, could be physical or virtual. Node represents the actual compute power.

In local, we only have one machine, so we're gonna have a single node cluster.

We need to run services in the cluster against your computer power.

To distribute the app's running in cluster across compute power, is done with another abstraction layer that's called a pod.
Pod is a containment of one to many applications that we wanna manage as a unit of compute.

The pods should be allowed to come up and down at anytime. So we're not gonna run anything here that's stateful. But sth like a DB
is stateful. So in an environment like GCP, we would probably have our DB out of cluster.

So in a production env, we're not gonna be running a pod in cluster for DB, we would be running DB outside of cluster and we need to
configure cluster so that our pods are able to access that DB that could be on a different network.

There are some DBs like DGraph that are designed to run in k8s cluster.

**Tip:** We want to be able to maintain, manage and debug the app.

`SIGTERM` is what k8s is gonna send for the shutdown.

In makefile's build command, we're setting the main package level variable named `build` to `VCS_REF`.

## 13-3.3: Write Basic Service for Testing


## 14-3.4.1: Docker Images
## 15-3.4.2: Kind Configuration
## 16-3.4.3: Core K8s Configuration
## 17-3.4.4: K8s Quotas / Patching
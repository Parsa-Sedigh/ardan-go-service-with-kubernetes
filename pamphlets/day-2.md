## 10-3.0 Intro
Deploy first mentality.

I want to get my local k8s env running and then eventually the staging or QA env up and running and CI/CD.

When it comes to go, think about layers not groups.

Average person can not keep track of more than 5 things in their head at any given time. This is why at the outer layer of the project, there are
5 layers. Layers that set guidelines and policies and provide things. A layer can have layers, but no more than five. It's good to keep this number
to 3.

You can even have layering inside a code file.

In service project we have 5 top layers:
- app: represents application-level concerns. All the different services and tools we're building, their main functions will exist in the `app` layer.
    - services: Doesn't have any layers inside of it because although we have folders nested, since inside those folders(metrics and sales-api) there
    is go code there(main.go), they are not considered a layer.
    - tooling
- business: any packages that we're building, related to business rules, business processing, anything specific that this project is trying to accomplish,
the problems this project is trying to solve will be in this layer. For our case, this layer has 5 layers:
  - core: core business packages
  - cview: core view packages
  - data
  - sys: system related packages. These are system packages but they're still very much tied to the business problem
  - web: web related packages tied to the business problem
- foundation: think of the foundation layer as the standard library for this project. These are packages that are not tied to the business problem.
Eventually these packages could live in their own repos and can be brought in as vendored packages. But for now, since we don't know enough,
we only need them in this one project, we put them under the foundation. We can call that future repo that the packages could be moved to in the future,
the **company kit** repo
- zarf: we need a layer of code that could hold everything that we needed related to docker, k8s, build, deployment. A lot of configuration and we need a 
name for this folder that would make this folder sit under `vendor` folder. Because we want that visual representation of different layers in the project.
What do I mean? We wanna see app layer above business layer. Not above foundation layer directly or vendor layer. So keeping these names alphabetical
gets tricky. So we needed a word that would sit under `vendor`. A zarf is a sleeve you put over a container so you don't burn yourself, like the ones in
coffe shops. Here, we also put container related configuration in the zarf folder!

Whenever there's go code inside of a folder, it's not considered a layer. For example in metrics folder we have main.go so it's not considered
a layer, it's a service(because it's inside the `service` folder).

Each layer gets to define it's own guidelines and policy.

An example of policy:

**Imports can only go down, they can't go up.** If I ever see a package in the business layer trying to import sth from the app layer,
we have a big problem. Imports can only go down. app can import anything, because it's application layer. Note: Inside the app layer,
we have different applications that we're building, in our case they are metrics and sales-api . We can't have imports between these as well.

Business can import anything inside business and foundation and vendor and zarf.

Note: In each layer itself we also try to keep packages separate(in app we can't import one thing from another app).

Packaging(**packages**) allows you to create firewalls between the different parts of your program. Our **layers** are also firewalls
between the different parts of our program. So when we say imports can only go down, it's important because it maintains the direction
that communication can happen between these firewalls.

**To create the project from scratch:** Create app layer. Create services layer. Add sales-api folder. Then main.go .

Q: What is the purpose of logs?

Logging is the window to the health of your services. Logging is the window to the problems that are happening. They are our first chance of
trying to identify what's wrong and fix it. You don't get to run debuggers in production.

So the purpose of the logs is to help me identify that the service is healthy and we can identify the bugs when they happen and be able to fix them.
It's for human-readable consumption.

But some people say: the logs for us is to record data. We don't use the DB for some of the data that we need to record, we put it into the logs.
So if we store data in the logs, you end up with 2 constraints there:
1. you need structured logging: because that's data, you have to parse it
2. what happens if you can't log for a moment for some reason? can you stop writing logs if it represents data? probably not which means your service
has to come to a halt. Now if your service is streaming television, you probably can't stop that stream just because you can't log.

Do we need structured logging? Can't we just use human-readable logs?

## 11-3.1: Tooling Installation
We're gonna use `Kind` for local k8s env.

Kustomize comes with the installation of kubectl. But we want to install kubectl separately because we're gonna use it separately.

You can install these with brew.

**Tip:** Choose a technology and use it not abstract it. It means for example when using uber's zap logger, we pass it everywhere as it is instead of
abstracting it so that one day if we wanted to switch libraries, we could do it maybe easily. If we wanted to maybe switch to zerolog, it's maybe
a day or less of refactoring and we rather do that in a day and not have abstraction layer. Get away from this idea that we have to abstract
everything. We don't! It's ok to commit to a technology. Commit to a cloud. It's gonna keep your code simpler. Every abstraction adds complexity.
**We don't design with interfaces, we discover them.** We develop concrete implementations first, we see how they work and then when it's necessary
we add that interface or that abstraction layer. And yes that means more refactoring.

Create a package `convenience`. Why we name it like that? Because we're not writing this package to abstract this app logger. We're writing this package
to make using it more convenient. We would write a lot of convenience packages. Inside there can be higher APIs. It's not an abstraction, it's convenience.

Create logger package under foundation.

### About packages
We want to write packages that provide(not contain). It should be clear what this package is providing.

Every package is an API. Everything you do in go is an API.

When you start writing packages that **contain**(which is not good), you're gonna have problems with naming conventions and dependency.
An example of package which contains is utils, helper, models, types.

Every line does one of three things: It's either allocating memory or reading that memory or writing to that memory.
In other words: you're actually only ever reading and writing integers. Unless you're working on GPU stuff then you're only reading and writing
floating points.

When we talk about integrity on micro level, you're saying that every allocation read and write, has to be accurate, consistent and efficient.

On macro level, we have the same thing. Every function you write is a data transformation(gets input, you transform that data and
returns output). Every API is a function, so every API is a data transformation. Any problem you've ever solved with code,
is solved through a data transformation. You're all transforming data all day.

So what is the purpose of a type system?

A type system provides 2 things: It allows input to come into the API and it allows output or data to come out.

A type system's job is to allow the input and output of data through an API.

So you should not ever have a common type system. Every API should have it's own type system. For the data that it needs and the data
that goes out. That maintains that firewall. There's no other dependency. remaining time: 21:00

## 12-3.2: Understanding Clusters, Nodes and Pods
Remember the deploy first mentality. So we need the k8s env up and running. Get that service as it is, running inside our cluster.

### K8s semantics
`Kind` stands for kubernetes in docker. So it's not even our machine that is providing the compute.
Docker is running inside our machine, we can think of docker as VM.

It is important to configure the docker to run with 4 cores and 6 GiB of memory.

The cluster we're gonna bring up will have 1 node with 4 cores and 6GiB of memory, that's fine for development env.

But our QA env might be driven across 3 nodes, each with 2 cores and certain amount of memory.

Tip: Nodes provide the compute power.

In our local env, we run the app in a pod, the DB in a pod, zipkin which is for telemetry and vault to handle secrets and
all of these run on separate pods, so that we can scale these pods individually. We can give these the services in the pods different resources against
the nodes that we have.

Running DB in a pod is fine in a local env, but in production, you would never want to run your DB inside k8s cluster.

One good thing about k8s is we can create a dev env that is close to our staging and production envs(aside running only one instance of anything
in local). So that if it's working locally on my machine, it's good chance that it would work on other envs.

To make sure you have the things:
```shell
kind version
kubectl version
kustomize version
```

The kustomize version that comes with kubectl is always behind, don't use that(you can see it's version when checking kubectl version).

To cleanup unnecessary things on docker
```shell
docker system prune
```

In order to bring up k8s is `kind create cluster` and we point it to the image and we give that cluster a name.

Tip: When working with k8s, we're gonna have different folders for our different environments. This is one of the nice things kustomize can provide.
Like a folder named `dev` and ... . There are a lot of magic tooling for this stuff like helm, knative, k9s which provides console based UIs to manage
things. But we wanna see the foundational things, that's why we stick to kubectl and kustomize.

`kubectl` has `wait` command where we can wait for certain things at certain conditions. For example in `dev-up-local` we have a wait command
since that is the last sort of service that gets created when we bring up our cluster, so we're just gonna wait for that to gets provisioned and then
we're gonna know that the cluster is ready to go.

We wanna see the cluster after it comes up. For that, we have `dev-status` in Makefile.

To bring up the cluster:
```shell
make dev-up
```

The first time you bring up your kind cluster, it could take a couple of minutes to do verification(`Ensuring node image` step can take a long
time for a new installation).
After previous command, run `make dev-status` in another terminal window.

So now we have a k8s cluster.

Now create the `docker` folder under `zarf`.

Instead of having one docker file that accepts all these arguments as placeholders, we rather hard code things(we see this in Makefiles as well),
just to make it simpler to read and debug. This is a choice you get to make for yourself.

Q: Why we're doing `ENV CGO_ENABLED 0` in dockerfile.service?

A: By default, CGO_ENABLED env var is set to 1. This means that the go compiler is allowed to invoke it's C compiler for any C libraries that are
being referenced during the build. But hey we're not referencing any C libraries. Sometimes you don't know this. For example if you wanna use
the race-detector, that requires C. So race detector won't work unless CGO_ENABLED is 1.

For example, when you're working on projects on dev and since you have CGO_ENABLED set to 1, we would be using C libraries and then when we're running
in container, we would probably set CGO_ENABLED to 0 and that would cause issues. This can make debugging hard, because you think you're
running pure Go and you're not. There are times when you have no idea if your dependencies are using any sort of C libraries or not and most of these
packages today have both options available depending on CGO_ENABLED is 1 or 0.

So in makefile, we set CGO_ENABLED to 0 when we do everything(like builds), to make sure that we're not attaching ourselves to anything related to the
C libraries and we also do this in the makefile when we run sth locally like tests. When you run tests, that's a build and you could be building against
C libraries you don't even know it.

Q: Do you want to vendor and if you wanna use the vendor folder when you do your builds, for let's say your deployment?

A: If you're not gonna vendor and/or you don't want to use the vendor folder when you builds your final binaries for deployment,
you then need to use these four commands in order.

If you want to use the vendor folder to not waste time to download the deps, we're gonna copy the project into the build image. Now I don't
want to copy everything! So we would create a .dockerignore .

If you keep the docker layers in place(which Bill never does, because he runs `docker system prune`), if the .go files don't change, the `go mod download`
command wouldn't run.

We can specify the .dockerignore file to list out folders and files we don't want copied over to build the image, when we do `COPY .` in dockerfile.
Note: This file is necessary for when you want to use the vendor folder because in this case you would do a simple COPY . , but we don't want
everything to be copied to building the image, so we use .dockerignore .

In `RUN go build -ldflags "-X main.build=${BUILD_REF}"` of `dockerfile.service` , we're setting a variable in the main package called `build`.
We can write the version and build reference in the metadata for the binaries(the DLLS and executables all had metadata in them).
This is the same idea. In the main package of sales-api, we have a variable called `main`. 
So if the value of build is "develop", it means we're running locally, we aren't in the scope of docker build.
So during the build, we can assign whatever we pass in as value of BUILD_REF to the build variable. For example you can pass in the commit sha to know
which version of code is being built.

You can do `make all` and we should have an image with our service in it.

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

------------------------------------------------------------
The base folder in k8s folder is base config that never changes regardless of the env you're running in and then we're gonna use kustomize to
help patch the base info to a local env like dev or staging or production.

**stuttering = redundancy**

We named the k8s namespace for our sales api `sales-system`. This name comes from `kube-system` namespace. We defined this namespace to avoid
everything being in the `default` namespace.

Pod names have an id attached to them and this id gets changed everytime the pod is recreated, so if you want to query the info about running pods,
instead of looking at the pods actual name, k8s lets us specify `labels` like:
```yaml
labels:
  app: sales
```
And by specifying the labels, we can query cluster with these labels and we wouldn't care what the name of the different objects are.

In `base-sales.yaml` , instead of having sth like: `ardanlabs/service/sales-api:0.0.1` as image of deployment, we wanna be more flexible.
Because in the dev env, we might have different names or different versions. Instead, we use service-image which is sort of a placeholder
for some config we're gonna add in the dev env and kustomize will be able to look up `service-image` and replace it with what we
know the service-image will be called in the dev environment.

In k8s/sales/kustomization.yaml , it's gonna tell kustomize what other files it should bring in to form the complete configuration that we need.

The file k8s/dev/dev-sales-deploy.yaml , is what we're adding to the base-sales deployment in (k8s/base/sales/base-sales.yaml). It is in dev
because that(dev-sales-deploy.yaml) is stuff that will be different in staging and production.

**Note:** The things in k8s/base/... never change. But things in k8s/dev are for development environment.

To patch the things in that k8s/dev/sales folder, we add the kustomization.yaml .

Kind can go to docker registry, but if we preload all the images, we don't need to go out to docker registry. So we're gonna use `kind load`
even for the images that we've pulled down which is good when you're on a plane that you don't have network and want to start the cluster.
So preloading allows us to we can run the cluster without a network connection to the internet to pull down the images.

**To apply changes to k8s(binaries(programs to run in k8s) has changed):**
```shell
make all # to build the docker image
docker images # to see the new ardanlabs/service/sales-api:0.0.1 image
kustomize build zarf/k8s/dev/sales # see the generated yaml by kustomize
make dev-load
make dev-status
make dev-apply
```
These commands will get automated!

If only the k8s config has changed:
```shell
make dev-apply
```

Q: Why k8s uses the word `apply` in commands?

A: When we want to apply the yaml files to k8s, it's not a synchronous thing(it's not instantaneous). Therefore we have `kubectl wait` command.

Note: You want to install kustomize and not use the one that comes with kubectl.

If you're running the app in docker, the `GOMAXPROCS` matches the number of cores associated with docker configuration.

**Note:** When you updated the code, run `make dev-update` so that the new image is built and is run into the local development cluster.
If the logs changed in the last code change, run `make dev-logs` to see the logs.

Run make dev-down to shutdown the k8s cluster. You can double check this by running `docker ps` to see if any related container of cluster
is running.

## 13-3.3: Write Basic Service for Testing

## 6-Kubernetes Quotas
- If we made a code change -> make dev-update
- if we made a config(k8s config) change -> make dev-apply
- if both -> make dev-update-apply

To get the environment up:
```shell
make dev-up
make dev-update-apply # if the source code has changed
make dev-logs # see the logs to double check everything
```

We want to run our cluster as closely as possible to staging and production environments.

Our cluster gets compute power from different nodes that are associated with the cluster and those nodes in essence are docker VM and in
our current config for development env, it's got 4cores and 6GiB of memory.

As an ops person you need to figure out how you wanna measure when your services running in pods are now at a point where we need to run on
**another node** where the CPU and memory is at a certain limit or size where it's time to scale horizontally. Note that we're not talking
about scaling vertically.

So in a production env maybe we start with one node and we're running some pods on that node and we get to a certain saturation point where
this node is getting to the threshold, so we wanna automagically throw another node and start running another set of instances of our pods
on those new nodes and when load comes down, we can get rid of some of the nodes because they cost money.

Part of a job of ops person is to figure out what strategy they wanna use to determine when we need to throw more nodes to run more replicas
and when to scale down and it's done through quotas through request and limit.

The `request` settings are used to determine where a pod could be instantiated on a particular node. So I'm(pod) requesting this amount of resource
and the system can say: Well I have a node that has that amount of resource, so we can run you on that node. So the `request` part
says: I'm requesting this amount of resource, is there a node available that can take me?

The `limit` value says: don't allow this service(container) to use more than this amount of resource.

One strategy is whatever we specify as `request`, is gonna match the `limit`. There are other strategies like where you ony set `request`s
and you don't set `limit`s.

Note: One thing we **don't** want to do is to not set `limit`s at all. Because if you don't set `limit`, you're allowing the each individual 
service on that node, to trying to take over the full resources of the machine all the time. This strategy could work with certain strategies
around auto-scaling, and it might not work.

Note: request and limit are not gonna be set in the `base` folder of k8s, because every environment(dev, staging, production) wants to set
these `request`s and `limit`s appropriately for that envionrment.

We don't care about memory. We're not gonna set any memory limits. We just focus on CPU limits for now.

**When we talk about CPU `limit`s, we're really not talking about cores. We're talking about time** and a lot of this idea of time,
comes directly from OS, linux. The way linux schedules work. So we're gonna generalize some of this stuff because depending on the OS
and their schedulers, there's cloud be technically little differences.

So when we talk about CPU limit, So we don't want to think in terms of cores, we wanna think of **time** and the basic time is **100 milliseconds**.
If we decide that we want a container to have full 100milliseconds(100 milliseconds/100 milliseconds), then essentially what we're saying is we wanna give this 
container 1 dedicated core. Because if it's gonna use the full 100 milliseconds every 100milliseconds, then it needs the whole core available
to itself all the time. Because we're never saying that it has to share anything.

But we might have a service like metrics service which we want it to have half the time(so 50 milliseconds/100 milliseconds). So for every
100 milliseconds interval, give this service half that interval(half of 1 core on every interval).

Now you might say, this service is so important, that I want to give it 2 cores of full time. What does that mean?

What it means is if I wanted to give it 2 full cores of time(2 dedicated cores), that's essentially 100 milliseconds of interval time on 2 different
cores. Then you have to start thinking in terms of 200 milliseconds. So we could say out of 200 milliseconds(100 on one core and 100 on the
other core) I want all 200(200 / 200). Now we're saying we want 2 dedicated cores.

When we say we want 150 milliseconds, we mean we need a core and a half. I need one dedicated core full time and take half of another one.

So this is a measurement of time on these 100 milliseconds intervals which ends up translating to cores.

Instead of milliseconds, we use the measurement of `m`. So 100m or 200m.

When you see 1000m, we're saying I want all 100 of 100. If we say 500m, we say we want 50 milliseconds of every 100 milliseconds.

This is why we stressed on make sure you have 4 cores for docker because we're gonna set `limit`s to 4 cores.

Each quota is gonna be different in different environments.

Note: Most of the time, the `request`s and `limit`s are rarely over 1000m, they are really small because the ops people would rather 
horizontally if that load reaches and give each service less resources.

If the `request` is bigger than what node currently has, that pod can't start on that node and if there are not any nodes that can handle it,
the pod won't start at all.

There is no concept of 1.5 cores. There's no concept of part of a core. So when we request for example 1.5 , we need to have at least 
2 cores available on our node to have this pod launch on the node. From go perspective, this:
```yaml
request:
  cpu: "1500M"
```
means we only need 2 OS threads to execute our goroutines. When it's:
```yaml
request:
  cpu: "500M"
```
it means we only need half the core which means we need at least one core of the node and we're dedicating myself as a single threaded(OS) program.

How the go scheduler works?

The magic of go scheduler is that it converts IO bound workloads at the Go level, to CPU bound workloads at the OS level.
The most efficient workload that you can run on your OS is a CPU-bound workload. Why? Because those workloads never get interrupted, they never
require the thread to be context switched. The Context switch on linux on good days costing you microsecond of time, it's like 12,000 instructions
that you've lost. So when you're running CPU-bound workloads where the threads are constantly busy and have no need to be context switched,
you're running much more efficiently, minus the OS having threads and interrupts and other things causing us a little bit of pain), from an
application point of view:

Let's say we have a 4 core machine and we're staring our go program on that machine. One of the things that the Go runtime is gonna do,
it's gonna ask the machine it's running on(physical or virtual): how many cores do I have available to do my work? In this case the machine
is gonna report back and say: you got 4 cores.

Now go runtime says: "since I'm really doing CPU-bound work, I'm gonna maximize my efficiency and I'm gonna ask the operating system for 4 threads
and what I'm gonna do is create an OS thread for every core that is available to me on that machine and I'm gonna create this concept of a
P(processor) which represents a core(a logical core). Then I'm gonna build these queues at each one of the processors and then I'm gonna start
creating goroutines which are application-level threads and I'm gonna distribute those application-level threads in these queues where I'm gonna
be able to run in parallel 1G for every OS thread that I have, because I have one per core(note: forget that the OS has threads, forget
the OS is involved for now), I get to have these OS threads locked onto the cores, they technically never have to move because they're not doing
any IO-bound work, they're doing CPU-bound work and what I'm doing at the go-level(application level) is my IO-bound work with the goroutines and the
processor."

So goroutines are coming on and off the OS thread which cost you about 200 nanoseconds as opposed to 1 microsecond for context switch(context
switch makes it 500 times slower!). So we don't want threads coming off of cores, we want goroutines come off of the threads. This could be
requests coming into the server and they're being distributed across these cores.

Note: Every go program is a CPU-bound program where the IO-bound work is happening at the go scheduler level.

So when go runtimes has 4 cores, that means it can run 4 goroutines in parallel at any given time and it'll keep those threads(OS threads) constantly busy.
Other than the OS needing a core, nobody else should affect me. I'm assuming that I'm the only go program running on that machine. Now if we run
a second go program on the same machine, they're gonna think: Oh Look I got 4 cores available to me and I'm gonna create 4 OS threads.
But obviously there's some context switching going on because two go programs are competing against each other for those cores. They themselves
don't know that there's another go program as well. This may or may not be a problem depending on how busy OS threads are. With IO-bound work
you can tend to wait more than actually do things. You're waiting on networks, DBs and ... . The beauty about go is that if you can pay for
garbage collection, you don't want to use any other language. If you can't pay for garbage collection(wait for that) this is where rust comes in.
We can pay for GC because we're building web services(we're working with network, so it's ok to wait for GC as well).

We don't want more OS threads than the number of cores. Because that's gonna automatically create context switches.

When `GOMAXPROCS` is 4, the app **thinks** it's a 4-threaded program(it has 4 cores available). Now when we set:
```yaml
request:
  cpu: "1500m"
```
What we're saying is we only need 1.5 dedicated cores. We need 2 cores actually, where we're gonna be using 1 core completely all the time and
half the other. So the go program in this case should not come up as a 4-threaded go program, it should come up as a 2-threaded go program(although
the `GOMAXPROCS` is 4, but we have specified a limit in the k8s deployment related to the pod of this program).

If we're only getting one and half the time, then we should only run with 2 threads on 2 cores. Why? Because we'll hit the thresholds a lot faster
if we're still running on 4 cores, because it's not about cores, it's really about **time**. We want to be the most efficient go program that we can
be, we wanna match our limits to the number of threads that our go program launches. That's how we can run at the most efficient.

Note: It's good to make our dev env to be as close to staging or production as possible, so whatever limits we're using on staging or production
we should put in the dev env too.

Q: What strategy should we use for request and limit?

A: It depends on what your program is doing? If your program is like a web service and it's steady stream, `limit`s are good. But if it's a program
that gets a burst of work, probably `limit`s are hurting you. Because as burst of work comes in, you eat that time quickly and then you have to
wait for the next 100 milliseconds before you can do anything else. So setting limits for programs that could get a burst of work is not good.
You want that explosion of usage to happen and you think about auto-scaling(hit the limit so put on another node) in a different way.

If you have a web API that gets a constant, fairly consistent work, then setting `limit`s is a nice strategy.

Note: The CPU limits are all time based and it's related to OS scheduler stuff.

**Note:** For every core in the system, there's a 100 milliseconds cycle. So if you say `cpu: 1000m`, what you're saying is you want the full
100ms / 100ms(you want the full 100 milliseconds of every 100 milliseconds cycle). Now if we have 2 cores in the system, we can represent that
as 200ms of time per interval. Now in that machine with 2 cores if you say `cpu: 1500m`, what you're saying is you want 100ms/100ms of core 1 plus
50ms/100ms of core 2. So 1500m is 1.5 core. `cpu: 1500m` means we essentially need 2 cores to do that.

**Note:** If you've just updated the k8s yaml files, run: `make dev-apply`.

Note: After adding some request and limit to reduce the number of cores(actually time) that your go program should use, if you see the log of
`GOMAXPROCS` of the program, you still would see the old value and therefore the program didn't pick up the new values related to request and limit.
This is bad. If the service thinks still has more cores available to it, it's gonna eat up that `limit` of `cpu` much faster.
For example we have set the limit to be `cpu: 1500m`, but the GOMAXPROCS is still 4 and not 2. Now how do we make sure that this go program
is only using 2 OS threads so that the GOMAXPROCS is set to 2?

A: There are few strategies for this. There's a package named auto max procs from uber.

To fix this, add this to base.sales.yaml:
```yaml
          env:
            - name: GOMAXPROCS
              valueFrom:
                resourceFieldRef:
                  resource: limits.cpu
```
Set GOMAXPROCS env to whatever the limits.cpu value is. But wait! GOMAXPROCS only accepts integers not strings like 1500m which is what we set
in limits.cpu . Under the hood, what k8s code for reading that value does, is it takes the limits.cpu, converts it into decimal point. So under
the hood, 1500m turns into 1.5 and then in their string function that's used to set the GOMAXPROCS, they use a math.Ceiling function and it rounds
it up to 2. 😀

So we can use the `<value>m` syntax with no issues here.

Now by putting it into base folder's definitions, you never have to worry about it again in any envionrment(dev, staging, prod).

Now we have a go program that is respecting the `limit`s.

To test this, look at the value of GOMAXPROCS by running make dev-logs, it should be 2 instead of 4. We're now running a 2-threaded go program.
Now we're running efficiently related to our limits. With this, the OS threads won't get context switches.

Note: If your ops people are using a strategy that doesn't require your go programs to be limited, then you don't necessarily have to do this(if
the `limit`s are not set).

**Note:** `request` determine where you can actually launch(start) that new pod with the specified `request` resources, but limit is to when to
detect the auto-scaling, **when** do we need to bring up a new similar pod, when the thresholds are hit to bring up another pod and scale
horizontally?

## 7-Configuration / Debug Endpoints
Q: When `GOMAXPROCS` is 2, does it mean we can only spin up 2 goroutines?

A: **No.** What that means is that the go scheduler is going to create 2 OS threads. but you can have unlimited(millions) of goroutines
executing, but only 2 of them can execute in **parallel**(at the same time), because you only have 2 OS threads.
There are local run queues at the P (processor) level which are maxed to 256 goroutines. There's also a global run queue(GRQ) that can have
an unlimited number of goroutines that are in queue waiting and as the local run queues have room for more goroutines, they pull from
the global run queue.

So your limitation is the number of goroutines that can execute in parallel, but not the total number of goroutines. But remember if
you're doing a lot of I/O bound work like a request going to hit the DB, that goroutine moves into a waiting state, so that thread becomes
available and it goes to run the next goroutine in the queue. **So for I/O bound work the GOMAXPROCS(how many goroutines can execute in parallel) doesn't
matter.** In fact there's a lot of go programs run single-threaded.

After considering logging, we need to think about configuration. Configuration rules:
1. the only place that should be reading any configuration(whether that config is in a file, whether it's envionrment variables or some
third party system like etcd) is in `main.go`. No other source code file, no other package is allowed to read configuration, otherwise it would become
chaotic. Configuration is only read at the beginning or startup and it's all read there(main.go) and then those **pieces** of configuration
are passed throughout the app **individually**. We do not create a master config type that is passed around program. We always want to have precision.
2. you should be able to type `help` in your service binary and see all the configuration options that are available including the defaults.
3. I'm not a fan of configuration files, they make it hard to handle secrets and ... . But it's mandatory that any defaults that you have,
should be allowed to be overwritten through an env var or command line flags. You need to support both, where the command line flag has rule over them all.
4. The default values need to work absolutely in your dev env. I wanna be able to clone this project and with little to no change, I should be
able to get it to run. Now, sometimes you need sth like an AWS key or ..., that has to be incredibly clear in the README or somewhere that they need that and
clear how they should add that to the project without ending up in the repo. But everything else needs to work on with their default value without
needing any additional work. The more your defaults work in your playground and production, the less it has to change throughout all the environments
that we're seeing in the yaml files.

Making things readable is about making things obvious. Simplicity is about hiding complexity.

To do readability and simplicity right means you write code that hides complexity without losing it's obviousness, that's a hard thing!

You always start with readability and then you refactor into simplicity, if you can find that.

It's good to prefix the environment variables with sth(like `SALES` in our case).

**Tip:** Always display the config that the service is using when it's starting. We're doing it in main.go in line where we have: `conf.String(&cfg)`.

What if there's a **config setting that is secret like aws key and we don't want it in the logs**. For that, we can use the `noprint` struct tag from the
conf package when we're defining that struct literal as our configuration. But Bill doesn't like `noprint` because we might think: Oh is there a bug?
Like that doesn't exist at all? And so instead, we can use the `mask` struct tag.

We want to maintain, manage and debug our services.

Q: What does it mean to be able to debug our service? Does it mean we can run a debugger?

A: Part of the logging is to help us debug our service. However, there's two more things we can add out of the box:
1. we can get full memory, cpu, blocking profiles and traces out of our running go program
2. go also has the metrics package in it's std lib

We wanna lay in the things from the std library that we need for profiling, tracing and metrics.

The `debug` package is business oriented and it's under `web` folder because it's only needed related to our web handlers.

**Security risk:** Too many developers in their main.go file, import this: `_ "net/http/pprof"`. Since the net/http/pprof package has an init function
they can import it like that for the side effects. What that import does is it says: Hey compiler, I'm not using the pprof package but there's
an init function that I need you to execute, so keep the import in there, although we're not using anything from that package.

Now there are some routes defined in the init function of that pprof package and by executing that init function, all of those routes are part of
the `DefaultServeMux` and then the developer is to bind the `http.DefaultServeMux` into `http.ListenAndServe`.

Q: Why is this dangerous?

A: Someone could do the same thing in his package that you're using in your project, without you knowing? He would add an init function
in his package that adds a route to the `http.DefaultServeMux` that exposes your service to very bad things. Especially if you forget to hide those
endpoints behind a firewall. With all the deps that you're bringin in your program, any of those deps could do this and by binding the
`http.DefaultServeMux` to server like `http.ListenAndServe`. So in the debug package, we create our own mux using `http.NewServeMux()` and bind that
mux to `http.ListenAndServe()` .

Now run the cluster using:
```shell
make run-local
```
and in browser go to `localhost:4000/debug/pprof`. With this, we're able to get memory profiles, cpu profiles, goroutine stack traces, traces, while
our program is running. So if the program is having problem, we can look at these.
You want the endpoints that exposes this info on separate IP or port, so you can hide them to just internal access, because someone could do
memory dumps or cpu dumps.

With `localhost:4000/debug/vars` we can see the metric data. But we don't want to read the metrics in it's default format. For this,
we can download `expvarmon`:
```shell
make metrics-view-local
```
With this, we have a terminal-based dashboard that's reading those metrics. I don't need to worry about connecting to datadog, prometheus or grafana.
We can read the metrics without them.

## 8-Cluster Access
**Tip:** Make sure the golang bin folder is in `$PATH`.

How can we deploy our service into k8s environment?

```shell
make dev-update
```

Now if you see `localhost:4000/debug/pprof` in the browser, it can't be reached. It's actually running in the cluster but we didn't expose that port.

How cool would it be if the debug service was a service that is supposed to just exist inside the cluster but we could still access it as if it was
inside the cluster even though we're outside of the cluster. This is what we want.

1. One approach is to use k8s nodePort or load balancer to access the app running in pod
2. another way is to use telepresence to create a tunnel(you can think of it as a VPN, but it's really not a VPN) and have telepresence run as
sort of a service(not k8s service but a an app) in our cluster and then telepresence will run kinda an agent service in the background outside of the
cluster(on our local machine) and with their API(CLI), it can set up a local network where the namespace for that pod could be accessed on some port. 
It means as we try to access that pod using any of our programs(like our browser which is not inside the cluster) they will go through
the telepresence agent which will talk to the TP service app inside the cluster and it will talk to the pod. Essentially using the same
namespace resolution that we have inside the cluster. 
This solution is good because what it means is I don't have to expose a port if I don't want to. So we don't open up 
anything but when we need to access some container in some pod, we just run telepresence. So there's no permanent hole(port) that's open to
our cluster just for some debugging or sth(you could expose your cluster though OFC).

Note that you could use your kube control instead of telepresence to just do port forwarding as you need it. There are also other products for this.
The beauty of telepresence is:
1. we don't have to automagically open up a port and have that reserved on my local machine
2. whatever namespace we're using inside the cluster, we can use it outside of the cluster as well which means that all of our defaults take the same
names and we don't have to change any of it in our config. Remember that we want to get the defaults work in every environment. A system like telerpesense
lets me use the same defaults while we're working on local cluster or ... .

We're gonna start working with telepresence before we open up any ports.

If you don't want to use telepresence you can just open up port.

Note on using telepresence on windows: telepresence is not able to proxy traffic from outside of wsl2. So the browser in windows won't work. So you have
to use curl or sth inside wsl2 environment. So if you try to open browser to hit endpoints of the app in the k8s cluster, the traffic won't get
forwarded through telepresence.

Note: You can run `make dev-status` in a separate terminal window.

After defining the clusterIP service, we still have not exposed the app to outside of the cluster, we can use telepresence to access the app.

Maybe this was a mistake by telepresence: When we run `telepresence connect`, will run it's agent(outside of cluster) and it wants to 
create three files at a location on disk that requires you to give it your password. Why they put them there? We don't know!
So when we run the `telepresence connect` if everything is successful, it's gonna ask you for your password .

After this command, we have a tunnel between the machine and cluster and we can send req to the apps. So we can access our app inside
the cluster without having to open up any ports, port forwarding.

We're a developer, we're trying to create a consistent environment with our defaults and we don't want to expose ports just for development that we
wouldn't expose on the staging or production.

**Open up ports approach:**

To expose ports all the way through local machine, add the ports in `k8s/kind-config.yaml`, but we need to restart the cluster to see the effect.

Stop the cluster first:
```shell
make dev-down
```

If you got telepresence working, run `make dev-up`. If you didn't get it working, run `make dev-up-local`

## 9-Shutdown and Load Shedding
All of the commands that have `-local` are for when you're not using telepresence.

Recap:
```shell
make dev-down # or if you don't use telepresence(in local development OFC): make dev-down-local
docker ps # make sure nothing is running(other than maybe other containers that is not related to this course)
make dev-up # or if you don't have telepresence: make dev-up-local
make dev-status # if you're running with telepresence, you should have the ambassador namespace and traffic-manager pod
make dev-update-apply
make test-endpoint # or make test-endpoint if you're not using telepresence
```
By using telepresence, it's like we're interacting inside the cluster which is kinda the best thing because our defaults can be used.

Note: if you just updated the code and therefore the binary changed, run: `make dev-update` .

We have launched a new goroutine for the debug and that goroutine has a mux and that mux is handling the traffic from the internet and anytime a
req comes in, it's(the debug endpoint goroutine) gonna launch a different goroutine to serve that traffic. These goroutines are specific to pprof or vars.
They're just reading data, they're not handling application state. So what we're doing in `Start Debug Service` section of main.go to create the
debug goroutine is not a good rule of thumb, we're really taking an exception there.

We have to create another goroutine for our application traffic. All of our application routes that handle the CRUD for our sales-api.
Now a req will come to the API and the application goroutine is gonna launch a goroutine for reqs.

After launching the application goroutine, the main goroutine has to be frozen, blocked, waiting for a signal to shutdown.

Note: The debug endpoints are just doing a read. But the application endpoints are doing more than reads. Some could be doing reads but others could be doing 
writes. This is important. Because if we get signal to shutdown, we can't just terminate the goroutines that are for endpoints that can do writes. We have
to wait until we know they're done. We have to implement **load shedding**. What we have to do is when we signal to terminate the program, we have to ask the
application goroutine(Ga) to shutdown, we have to wait for that goroutine to see that all of the goroutines that it had in flight are finished. Because if we just
return the main goroutine, we just let the main function return, then the go runtime is just gonna killing those goroutines spawned by the api goroutine and they could be
in the middle of a write, so we might have data corruption.

We don't care if the go runtime kills the debug endpoint goroutines, because they're only doing reads. We don't need to do any load shedding for those goroutines.
**Since we don't need any load shedding there, we're gonna create a goroutine that blocks on `http.ListenAndServe()` and when the main function returns,
anything that's in flight just get destroyed.** The goroutines that just do reading, will never corrupt the application state or data if they just
get destroyed in the middle of their work, so we can keep that piece of code simple without doing load shedding. But we can't do the same for
application-level endpoints that spawn goroutines. We need load shedding for those and the http package provides this.
![](img/9-1.png)

- GM is main goroutine. It spawns GD and GA goroutines. But it's gonna block until a terminating signal is received.
- GD is the debug server goroutine
- GA is the application server goroutine
- that cloud thing at the top is the mux of the debug endpoints

### Concurrency and parallelism
**Concurrency means undefined out-of-order execution.** An example of undefined behavior: map iteration in go is undefined. The spec has not defined that behavior.
The current go compiler implementation for map iteration once the map gets to a certain size, is random.

Now since the spec does not define this, is the go team allowed to change the map iteration order? They are legally allowed to do it because it won't break the
backwards compatibility promise because it's not in the spec.

So if you have a program today that relies on the fact that a map iteration order is random, your program could potentially break.

Out of order execution: It means we can have a system with a single thread on a single core with thousands of goroutines that are being context switched
on the P. What is the order of these goroutines execution on that M? From concurrency perspective it's undefined out of order. This is why
concurrency can exist even in a single-threaded environment, because it's about out-of-order execution.
If your algos don't support sth being done out of order, you shouldn't be using concurrency. You should use one goroutine and make sure everything
runs sequentially.

parallelism means we're gonna have two or more goroutines executing exactly at the same time. You can only get parallelism if we have 
two cores with two Ms or more.

**Note:** If we have only one core, we can still serve hundreds of thousands of reqs, because of concurrency and out of order execution that can happen.

The main goroutine(GM) is the parent of all, it's the first goroutine that exist.

As a general rule: The parent goroutine should not terminate before it's children goroutines. When the parent goroutine is terminating, 
it's because all of it's children goroutines have terminated. You want solid, no goroutine leak code. But there are times when the parent
goroutine does have to terminate before it's children. There are times that this has to happen.
But if you're cocky, you would have orphan goroutines. Because the moment a goroutine is orphan, you can not have a controlled shutdown. Because
when we shutdown the program when a terminating signal is received and that orphan goroutine is in the middle of a write, we would get corrupt data.
A case to avoid: People inside of a handler launches another goroutine and then the handler returns(the parent goroutine terminated) and the child
goroutine is now an orphan, we don't even know it exists, it's still running, now we shutdown the service and the orphan also gets destroyed by the
go runtime but maybe it was in the middle of a write and therefore we would have a data corruption.

But wait. As we mentioned, but happens when the parent goroutine like in a web service, has to terminate but you want to do that 
extra work asynchronously in the child goroutine? That orphan goroutine needs to be adopted! If you're a parent goroutine and you're gonna launch
a child and you're gonna terminate, somebody has to adopt the child. 99 out of 100 times, it's the main goroutine that's adopting that
child. How do we do things like that? At a high level, the parent goroutine could be using some sort of package that keeps a list of these goroutines,
with a shutdown function. The main goroutine knows at the end of the main function, that it has to call the shutdown function of that package and that
function will do a wait for the children(adopted goroutines that are in the list). That's a way of finding a foster parent for the orphan goroutines.
We do it by making sure that there is some package that has state that has knowledge of these children with some sort of shutdown api(in our case
a method), so that the main goroutine can adopt these children (since their parent are gone) and still make sure that those children
terminate before the main goroutine terminates.

Anytime looking at concurrent code -> who's the parent? who's the child? Is the parent terminating before the child? If yes, Oh oh, why is
that happening? Is it necessary? If not, don't do that. But if the parent has to terminate before the child -> who's taking ownership of
these kids? Does it have a way of knowing they still exist? How does it know to wait to shut them down? How does it know how to signal them to shutdown?

An example of orphan goroutine in our case is the goroutine that spawns the debug server! The parent of it is the main goroutine, but it loses
track of it child. It has no knowledge of the child existence. But we don't need to wait for that goroutine or it's children because
all it's doing is reads(all the endpoints are reads) and we don't care it gets terminated in the middle of reading. But we need a way of **waiting**
for all of it's children to terminate and for the parent goroutine(responsible for running the application API) to terminate, before the main goroutine
to terminates.

Wait groups and channels create orchestration. That's two or more goroutines interacting with each other.
A lot of times, a wait-group can give you orchestration without the complexity of a channel.

The other thing is synchronization. Trying to access shared state in a synchronized way. That's where atomic instructions or mutexes come in.

Sometimes you don't need orchestration, you need synchronized access to shared state.

Q: Can bubble sort be done in a concurrent way? Can you create multiple goroutines

A: No. It can't be done concurrently. There are algos that can't be done in an out of order fashion(concurrently).

**Channels serve one purpose and that's for signaling.** Horizontal signaling. Do not think of a channel as a data structure. With channels,
one goroutine can signal another goroutine with or without data, some event.

If the **signal** doesn't make sense in the context you want to use it, you should not be using a channel.

There's two ways you can signal in go. You can signal with a guarantee or signal without guarantee. Guarantee means we can signal where the sender
has a guarantee that the receiver has received it. For that to work, the receive has to happen nanoseconds before the send. This is where
unbuffered channels come in. unbuffered channels give the sending goroutine a guarantee that the signal being sent has been received. The cost of this,
which we have to pay, is latency. If the receiver is not there, the sender has to wait.

So there's a cost of latency for the guarantee. Nothing is free!

If you don't want the guarantee, then you use the buffered channel. In buffered channel, the send happens before the receive. If there's
a buffer space, the send can complete. The latency is less than buffered channels. But we're losing the guarantee, you don't know if that signal
has been picked up by the receiver. You don't know when it's going to be picked up.

There are signaling problems where you need the guarantee and there's signaling problems where you don't care.

**Anytime I see an API(method or function) that accepts or returns channels, it is a smell.** But if you have a legit reason to use a channel
as part of your API(in or out), you have to decide who needs to dictate the guarantee or not. Do I get to dictate it as the API itself?
Or does my user need to decide that?

In case of `signal.Notify()` of std lib, it has decided that the user should dictate if they need the guarantee or not. We as the user,
decided we don't need the guarantee because we made a buffered channel with capacity of 1. Why? Because these are signals. What happens if we
hold down ctrl+c? How many signals am I creating every second? Thousands of them. Should the goroutine behind the that's wrapping the
signal.Notify() block because we didn't receive that? No. if the receiver is not there to receive a signal, it drops it. We only care about
one of the signals, we don't care about the rest. So the API design is good. Because the signal.Notify() has to let the user decide.

For example in the serverErrors channel is buffered, so that when we receive an error from api.ListenAndServe() , it won't block,
and the goroutine terminates. If it was unbuffered, the goroutine would never exit and we would have a leak. In other words,
api.ListenAndServe() wants to send to the receiver, but there's no receiver and if the channel is unbuffered, it's gonna block forever, we would have
a leak. But we're using a buffered channel of 1, we allow the goroutine to terminate fast in case of error(without any blocking) and we pick up
the signal when we need it.

So for creating an http server, we launch a goroutine, it blocks on api.ListenAndServe() , however if it fails, it's gonna return an error and we 
send that error(through signaling) into serverErrors chan and we let the goroutine to terminate in case of error. Then in the `select` we block.
A `select` is a blocking call. You wanna block a goroutine forever? Do this inside the goroutine function body:
```go
select{}
```
That's it, this blocks the parent goroutine forever.

Q: How do we know the timeout value for load shedding?

A: We have no idea, we just put a value!

**Tip:** We can add the todos(list of things that we have to get back to) of the project in the main.go right after the imports. Two benefits are:
1. we can move on to whatever else we're doing actually, we're not worry about forgetting it
2. when you have some time to do a thing, we can look at this list and see if there's anything we can do in our time and if that todo was a big thing
and required more time, we can add a jira ticket for it

**Tip:** We can have a big run function in our main.go instead of encapsulating pieces of code in there into functions, this will be just so much easier
to read and manage.

The command:
```shell
make dev-restart
```
Should send a terminating signal. Approve this by running: `make dev-logs`.

Now we need a mux to define the routes and bind them to handler functions.

## 10-Application Mux
### Web framework
**Polymorphism means a piece of code changes it's behavior depending upon the concrete data it is operating on.** We're able to do that
by allowing the input of concrete data to be passed through it's behavior not based on what it is but based on what it can do.

The mux we're gonna use is `httptreemux`.

```shell
make dev-update
make test-endpoint # or make test-endpoint-local if not using telepresence
```

An API is about data transformation. Data in, data out. When it comes to passing data into an API, you have two choices:
1. you can accept that input based on what it is(concrete type of data)
2. or you can accept that input based on what that data does(interface type of data)

**But you should always return the data out using the concrete type.** Do not pre-decouple data, you're not helping anybody. The decision to
decouple sth is not your decision as the API designer, it is the caller's decision. You(as the API designer) get to decide if you want to accept
concrete type in a decoupled state, but you don't decide for the caller with the data that goes out.

But we're breaking our rule in `handlers.go` APIMux func. We're returning an interface named `http.Handler`. We instead return a concrete type named
`*httptreemux.ContextMux`. Nothing breaks in the main.go because `*httptreemux.ContextMux` implements http.Handler interface.

**Note:** The only time it is ok to return an interface, is when we return `any` interface, because you're trying to do sth really general.
But we've go generics in go. So use generics instead of `any` or `interface{}`.

**Do not use an interface as the return type.**

Note on generics: It all comes back down to **polymorphism**. The interface type allows us to write APIs that are polymorphic.
Let's say we have an API that says: I will accept any piece of concrete data that conforms to this **behavior(interface)**.
But we call that polymorphism with interface: **runtime polymorphism**. Because we do not recognize the behavior of that code until we run.
It's only at runtime when we know how that data is gonna behave.

Now generics in go is really not that much different. A generic function in go is a polymorphic function. In those functions,
you're deciding what the type parameters for that input and potentially output, are going to be. The difference is that generics lets you write
polymorphic functions that are compile time based not runtime. In other words, we know at compile-time what the behavior is gonna be not at runtime.

So a generic function is a polymorphic function where we determine at compile time what the behavior of that function is gonna be.

Go had generic functions before they introduced generics! Like make, append, new, close, delete. All of the built-in funcs are technically generic funcs.
They can accept values of different types. The difference between these and the funcs with new generics feature is only the go team could write them!
You couldn't write a generic func yourself.

If you're gonna add generics to your programming lang, which means adding compile-time polymorphic funcs, you've got two choices at compilation:
1. write a physical concrete implementation of every function for every type that you know it's gonna be used for that function. The advantage
of writing an individual concrete function for every type is those functions are gonna run like concrete functions. Very fast and probably in many
cases, zero allocation. The drawback is compile times take longer depending on the number of types that are gonna use that function(we have to write
that function for every type)
2. instead of writing a separate concrete func for generic function for every type that uses that generic func, they write one func that will
be used for all types(the same tactic they used for built-in funcs). The advantage is we get quick compile times, but is the downside is that
generic funcs are slower? No(in case of go), the generic funcs run at the same speed as runtime polymorphic funcs

One of the strengths of go is fast compile time. The language team won't do anything to make it slower. Go uses approach 2.

**Note:** Use runtime polymorphic funcs until you can't and the only use case for when you can't use runtime polymorphic funcs is when you wanna return.
When you would have to use an empty interface to return, or you're building container types which are data structures, linked lists, stacks, queues,
maps and ... , that's where the generics are really gonna shine.

These two types of polymorphic funcs(runtime and compile time) have no difference in performance!

So again: We don't want to write funcs that use interfaces as return types and if it ends up being returning an empty interface return type,
we now have generics. So at compile time we can use the concrete type(the type parameter) instead of an empty interface.

We're gonna have groups of handlers for the different domains.

Under the handlers folder, we have a folder for each version of our API and under each, we would have a folder for each group of handlers.

Q: Isn't that(like homegrp) a containment package? that contains all those handlers?

A: Well it's containing handlers related to that domain.

What is the policy that we wanna set on handlers? What should a handler do and what shouldn't it do, to make sure we have consistency.

Frontend dev should participate in the design of web APIs.

It's OK to have different result models(2xx), but different error models means you can't generalize your error handling and that's very bad!

We want to put boilerplate code around logging and error handling, so that they are consistent.

All a handler does, is 3 things:
1. validate the input(if there's any) that's coming in
2. call into the business layer to do any processing
3. return errors(if there are any), note that we don't handle errors in the handler, or handle the ok response if there is no error

But there's a challenge and every mux is gonna present this challenge.

We have a goroutine that's listening on a `net.Listener` for some HTTP traffic that's coming in. The job the mux is to look at the req
select the right handler based on the ServeHTTP func and launch a goroutine to do the work. The job of that goroutine is to call the possible
middlewares and the handler.

The handlers layer is kinda the outmost layer, yes mux is the outmost but it doesn't know anything about business processing. This is a big
problem because if I wanna return errors from handler to sth, I need sth in between mux and the handlers. I don't have the ability to do that.

Every mux is gonna treat my handler func as the most outer layer.

We want to wrap the handlers layer like an onion. Each layer does one thing. So the layers are like(from outmost to inner layers):
- log
- error
- handling panics
- handlers
- ...

With this, we can return an error from handlers layer and that error can be process all the way out. So there are some layers before the handler and they
execute code before and after the handlers layer. We can't do that with our router. So we build our own router! **We want middlewares.**

We need a way of doing this. We want to also pass a context to the handler funcs and also return possible errors from the handlers to outer layers.
We want a handler func to look like:
```go
func Handler(ctx context.Context, w http.ResponseWriter, r *http.Request) error {}
```

If the context package was in the std lib from day 1(from v 1.0), the handler functions would look like the line above.

From an API perspective, you want your context to be the first param of any API that does any sort of I/O or work that may need to be cancelled.
Also if you like tracing like open telemetry, then you need that context, so you can gather that info throughout the call chain.

**General rule:**
- Foundation-level functions should never need sth from the context in order to run.
- no business layer func should need sth in the context in order to succeed. Note: Business level APIs can be used for all sorts of different apps:
CLI tools, web apps. If suddenly you need to hide a DB connection in the ctx for a business api to work, how does that CLI developer know that?
We can pull those things out of the ctx at the app layer, pass them into the business layer.
- in the app layer we can put things in the ctx, but the handler has to pull them out(?).

**Do not hide things in the ctx.** Things like loggers, DB connections and ... which would cause these funcs to fail if sth is not in the ctx.

But let's say there is a legacy software and in prod, but we don't have any logging. If ctx is used in the software, instead of 
breaking everything by passing logger everywhere, we can break the general rules and hide the logger in the context. But you have to be careful that
the logger is in the ctx when the req came(we don't know about all of the source code), so we have to check if the logger is there.

## 11-
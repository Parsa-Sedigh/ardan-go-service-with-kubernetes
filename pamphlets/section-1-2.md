## 1-1.0 Intro
## 2-1.1: Design Philosophy, Guidelines, What to Expect
## 3-1.2: Tooling to Install
## 4-2.0 Intro

## 5-2.1: Adding Dependencies
What is a project?

A project is a repo of code.

Don't think every **application project** can only be managing one binary.

A project gets to define the design philosophies, policies and guidelines that need to be followed for that codebase.

By creating a go.mod , it allows the go tooling to understand: This is a go project initialized for modules and this is the root folder.
`go mod init <name>` needs a name because the compiler needs to have a namespace for everything in order for it to be able to locate the
right code.


What's beautiful about go is don't share binaries, share source code.

**Tip:** For the name of module, use the name of the repo. Nobody else can have github.com/<your username>/<repo name>.
So for name of the module, we can for example use: `github.com/Parsa-Sedigh/<repo-name>`
Nice thing about using a URL as namespace(actually name of our module) is it's gonna give go tooling another opportunity to be able to
fetch code for us. Because it's already gonna have the unique url of where this code exists our there on the public or in case of
private repos, on the private network.

When we see go 1.17 , it's telling us this project is compatible with everything related to go 1.17 and **beyond** but we can't assume
that it will be compatible with earlier versions of go.

This version also does one other thing: the tooling tries to maintain backwards compatibility with defaults.
This way as go releases new versions, your build scripts won't break.

You shouldn't have multiple go.mod files in your project.

There are some exceptions: We maintain 10 different 3rd party dependencies, we don't want 10 different repos. So at the root of repo,
we create multiple folders and each folder has their own go.mod .

But if your project is an app or one 3rd party dep, use one go.mod at root.

```shell
`go env`
```
`GOMODCACHE` env variable points to your module cache and your module cache is where all of the source code for any 3rd party deps that you're
downloading goes, so go compiler can build your source code against that source code.

In `/Users/parsa/go/pkg/`, you **can** run:
```shell
go clean -modcache
rm -rf sumdb
```
To remove the module cache. We're now completely clean, there's no go 3rd party dep source code in our computer.

**go please** is the language server that the editor talks to for intellisense and errors and ... . Now when the language server
starts up(it starts up with the editor), it stores in memory, a cache of the module cache. So if you really wanna start clean after
running the go clean and removing sumdb, you need to restart the editor to shut down the language server, because it already had
it's own cache.

We need to get the 3rd part source code on disk:
```shell
go mod tidy
```
This command walks through your project, looking to validate that all of the source code your project needs, is on disk in the module cache.

## 6-2.2: Module Mirrors
**Q:** What happens when we run `go mod tidy`?

`GOPROXY` env var(you can see it's value with `go env`). Sth like: GOPROXY="http://proxy.golang.org,direct". This env var, directs the tooling
in where to look for the source code it needs to download and by default, the value is the proxy server(**module mirror**) hosted by 
the go team(http://proxy.golang.org) and if that proxy server can't find the code that it's looking for, then it will go `direct` to the
version control system. `direct` could be github or gitlab or ... (what we specified for the url of package).

So by default, when downloading packages, go tooling will hit the google proxy server first and if it can't be found there, then go
`direct`.

The job of proxy server is to proxy all of the reqs for version control systems In other words, when we wanna download source code 
on github, gitlab ..., the proxy server proxy those reqs to version control systems so we don't have to constantly go directly to
those sites, we can go to only one place(proxy server) and get what we need. There are also other benefits with the proxy server.


Since we don't have the source code already in module cache, so `go mod tidy` goes to talk to the proxy server and it asks proxy server: 
What versions of this package you know about? and in response, the proxy server returns a list of versions.

Now where does this list of versions come from?

Everytime someone asks the proxy server for a specific version of a package(repo) and the proxy server doesn't have that, the proxy server
goes to a process of saying: Let me go see if that version exists on github and if it does, it then generates **a module of code** for that
repo(the user asked) at that version. It do this by pulling all of the code out of repo for that tag the user wanted and then created a 
zip file labeled as: `<package name>@<version>`. So a module of code would be a snapshot of all of the code in a repo at 
some particular tagged version.

So inside the proxy server, we would have a zip file named `github.com/<username>/<package name>@<version user wants>`.

Each of the tags in git, represent a module of code at a given version.

So when we ask for package with version, proxy server gonna look at it's catalog of packages with their versions(if they exist, if not,
will pull that version) and then it sends back a list and then `go mod tidy` has to select the version and (if no version was specified)
it will choose the latest. Now go mod tidy sends another req and says: Give me the package of <package>@<version> (which is based on the list
that the proxy server sent us before). Then the proxy server sends back a zip file.

Now let's say we were **directly** asking for a version of package and it wasn't already in the proxy server's catalog, proxy server will
go out and look for, it will create a new zip and sends it to us.

After we got the zip file, it's unzipped in the module cache.

The go.mod will also be updated. It will list the module(repo) that has the source code we want with the tagged version that was selected.

A go.sum file was also created(or updated) and this file stores hash codes that allow us to validate that the source code that we got, is the
source code we should've expected we get.

Now if the proxy server returns that it doesn't know anything about the package we requested, then go **direct**. So go to github, gitlab
or ... directly and see if you can pull it yourself.

Some people have a problem with these defaults:
- Maybe it's because of privacy because we're requesting sth from google's server.
- the teams is not using public repos. They're using private ones. For example they run their own private version control system. Maybe
they're running their own gitlab(maybe running in a private network). So when we would have private repos on github or gitlab(not on a private
network) or ... or running our own instance of gitlab(in a private network), the proxy server can never access that stuff.

There are 2 private repos here:
- private repos on public sites
- private VCS(private network, running private gitlab yourself)

So you never wanna go to the proxy server for the private code, since it will always return a 404. This is where `direct` comes in.

Why go to proxy server to always get a 404, when we know there's no way it ever going to find it?

In order to deal with private repos, there's another set of env vars: `GONOPROXY` and `GONOSUMDB`.

`GONOPROXY` allows us to put in the domains that we don't want to go to proxy server for them and instead would go `direct` for them(if
you have private VCS).

There is another option that we can go to fix this problem:

The `proxy.golang.org` specified in `GOPROXY` isn't the only proxy server that exists. There are a couple of other proxy servers:
- athens: Open source proxy server. You can run your own proxy server on your own private network. Then you would change GOPROXY to not to
point to proxy.golang.org, but to `<your proxy server>, direct`. You should also configure athens to say: I'm still OK with you going
to proxy.golang.org or ... .
- JFrog's artifactory: if you're using Jfrog already, it has a product called artifactory and it has a proxy server builtin(you wouldn't
buy artifactory just for the module mirror, instead if you already using it, you can set it up).

`direct` is gonna be slower than `proxy.golang.org`.

## 7-2.3: Checksum Database
Go tooling generates 2 hash codes. One for all of the source code unzipped and then one for go.mod . Then the tooling first asked: Am I
writing these hash codes to go.sum for the first time?

If yes **and** we have an env var `GONOSUMDB` is either empty or not the same as GOPROXY domains, there would be another roundtrip to
checksum db which runs by the go team. Everytime proxy server creates a new zip file for a module with a specific version, it generates
the two hash codes and write those hash codes to that checksum db. So for every module with a specific version, there are related hash codes
in checksum db. Now when we download the zip file from proxy server, unzipped it, added to module cache, wrote to the go.mod and generated
the hash code and then we asked the checksum db for the hash codes of the package we downloaded and then we compare them with the hash codes
we generated, if they're the same, we're done, the hash code matches the hash code in checksum db then we know we have the same code
that proxy server saw for the first time.

Now since these codes are already in the go.sum , then we don't do this extra trip to checksum db.

Everytime we download the package with the same version, the hash code should be the same. How could it not be the same?

If we don't go to proxy server and instead go direct, or go to private VCS or private repo. But we always make sure the code is the same
with go.sum hash codes. 

So no matter where you pulled this package from, we can make sure it's the same exact code and nobody could remove the tag, change some code and
then put the tag back and you get malicious code.

If you're using the proxy server, then you get the benefit of also making sure that the code you're pulling no matter from 
where, is always the same.

This is why go.sum and go.mod have to be part of the project.

**Note:** tidy command is putting code into module cache.

## 8-2.4: Vendoring
Vendoring means not to use the source code sitting in the module cache, but to bring all that 3rd party dependency code into the project,
so the project owns every line of code it needs to build. The source code stays in module cache and now also in our project.

You should be vendoring until it's not reasonable to do so. When it's not practical? When the amount of code you're managing in that `vendor`
folder becomes huge.

```shell
go mod vendor
```
So now the build tooling won't look at the module cache, it's gonna look inside the vendor folder.

Note: You need to push the vendor folder up.

### Benefits of vendoring
- we have access to code of 3rd party
- we can debug the 3rd party code and change it
- we own every line of code
- we can sleep at night knowing that that version of your project can be built everytime because it doesn't have to go anywhere else to get
the code it needs to do the build 
- we can also remove one package from vendor and it will be removed on repo after pushing. Now anybody trying to go `direct` to download
our repo, their build will break. But let's say they get it from proxy server instead of direct. Well in that case, we tell google to remove
it. But still they could have it on their own **proxy** server(like Athens).

These are the benefits that we don't have when we're working with the module cache

## 9-2.5: MVS Algorithm
minimum version selection algo.

To install a new package, you can add it to imports then call a imaginary function of it like New() so the tooling won't remove it and
then run:
```shell
go mod tidy
go mod vendor
```
When we installed the `github.com/dimfeld/httptreemux` package, we got `v5.0.1+incompatible` in go.mod .

Also you see the latest tag of that package currently is 5.3.0 but we got v5.0.1 !

v5.0.1 doesn't have a go.mod and the version after it v5.0.2 added a go.mod . Why?

First in the go.mod of that package, we see it has a /v5 at the end of it's module name. What does this mean?

Go supports the idea that you can build against different major versions of a module. When there's no `/v<x>`,
then we assume that we're talking about v0 or v1 , so we don't have to add that for kinda backwards compatibility.
So in source code, when we import a module without `v<x>` at the end, we're telling the tooling we want the latest version of **version 1**!
Not version 5 and since we have a situation where v5.0.1 of module doesn't have module support(no go.mod), the tooling will treat that
as the latest version of v1, because it has no module support.

So how do we get the tooling to give us the actual latest version?

You have to match the namespace in the module. So we need to add that v5 when we import that module in source code. So it won't download
the latest of version 1(yeah this is correct), instead it will download the latest of version 5.

So run tidy and vendor commands after adding v5 to import.

Now we got `v5.5.0` instead of `v5.0.1+incompatible`.

So one of the things we need to do when installing a 3rd party package, is go to the repo of it and look at the tags and if it's going beyond
`v1...`, then we wanna go to the latest tag and validate that the developer is using the right naming convention(namespace) for that
version.

### MVS algo
Let's say our app has direct dep on package A v1.2 and package A itself has direct dep on package D 1.3.2 . But the latest version of D
is 1.9.1 . Now what version of D the build tool should choose? 

It has 2 choices. It can choose 1.9.1 or 1.3.2 . If you're using traditional version selection tools which use algorithms like 
SAT solvers, SAT solvers want to pick the latest version of D. It's gotta be the best version.

The problem is that is not necessarily what go's tooling gonna do. Because it uses an algo called MVS. Go's tooling says: v1.9.1
isn't necessarily the most stable code for building this app. Because A 1.2 is reporting that is compatible with D v1.3.2 .
Note: Remember version semantics is a social contract. We're **assuming** v1.9.1 relative to v1.3.2 , the api didn't change. In other words,
you're **assuming** A can build against D v1.9.1 and the build wouldn't break. But that's an **assumption**. But A is guaranteed to be built
against v1.3.2 because it compiled with it to be available to people, right? It **guaranteed**.

So the MVS algo wants the guarantee, so it chooses D v1.3.2 .

Let's say our app brings package B with v 1.4 and B also uses D but it's v1.5.2 .

Now what version of D is gonna be selected?

It'll choose v1.5.2 and if we were to no longer use package B, the build tooling would still build the app against D v1.5.2 !

Let's say we want to use the latest version of D here. To do this, we can use go get with some flags and in this case, it begins to walk the
project tree and we're telling it to update every single dep to it's latest.
```shell
go get -u -t -d -v ./..
```
Then you can run the tidy and vendor commands.

So we can use `go get` for updating commands. You can also use it to only update your **direct** deps and letting MVS play out for indirect
deps.

Go also solves the dimond dependency issue:

The scenario is A using D v1 and B is using D v2 (if they were the same major version, we wouldn't have a problem).
We can't assume B can build against D v1 and we can't assume A can build against D v2 .

To fix this, we can bring the both versions in go.mod and both will be stored on disk in different folders.
## 11-Web Framework / Middleware
We use value semantics(not a pointer) for data but if a type represents an API, we use it as a pointer, because it should be shared. 
Any state related to this API usually shouldn't be copied. Like the App type in the `web` package.

When we embed a struct into a parent, the parent also would implement the interfaces that the the embedded struct, implements.
```go
package main

type B struct {
	// ...
}

type A struct {
	B
}
```
A also implements the interfaces that B implements. We don't have inheritances or sub-typing in go.

The idea of embedding is not about sharing state, **it's about sharing the behavior.**

To create that onion layers(every layer is a middleware), we can put any code we want to run before calling the handler, simply before calling 
it in Handle method in web.go and we can put any code we want to run after handlers, simply after the handler call! So we're essentially wrapping 
code around the handler.
**We did this by embedding the `*httptreemux.ContextMux` into our own struct(App) and overriding the `Handle` method of the interface that it implements.**

We don't want to do the logging of the req in the handler. Because we would have 100s of handlers, we don't want to do the same thing 100 times. We want
to use a middleware.

There are middlewares in the `App`(mw field) like loggers and middlewares in one specific handler, like auth middlewares.

Note: We can't import code from one foundation package into another or from layers above foundation into the foundation.
For example, web package can't import the logger package(both are in foundation).

If you had a package that doesn't make sense to have a file named after the package name, it's a **containment** package. For example the `mid` package
in our project's business layer. It contains all the different middleware functions. And therefore this package is not ideal. But remember you can
get away with a little bit of containment in the business layer, but never do that in the foundation layer.

We usually put a couple of lines above the `package x` for an overview of the package. But if this overview is large, create a doc.go and put all the
overview comments there above `package x`.

The indentation will determine if it generates code blocks or text.

Ideally the file with the same name as the package would have the overview comments. So in package x -> x.go would have the comments. But if those
comments were a lot, put them in doc.go instead of x.go .

Q: How do we pass the logger to mid.Logger() func?

A: We can't pass it into that middleware func, it shouldn't accept the logger, that's too big of a policy and also our handle doesn't accept it because
of the same reason.
We have two choices to solve this problem:
1. Not a good approach: define a struct that has logger as a field in the logger package and make the Logger func to become a method for that type. Now we have the
logger through the method receiver. We don't like this for this situation because we're asking the app layer to construct the type before it can
pass the logger to the business's web framework. We shouldn't need to construct sth in memory in order to gain access to this behavior.
2. make mid.Logger returns a middleware instead of a Handler. The middleware itself returns a handler. We don't need a new type and 
a method for that type, we can write a function that that accepts a logger returns a function.

**Tip:** If a func wants to accept 0 to unlimited number of args, use a variadic, but if it wants to accept 1 to unlimited number of args, accept a slice.

## 12-Tracing / Error Handling
We don't know which `started` is for which `completed` log. We don't have a trace id for each req. We wanna generate a unique trace id that we can
use in the logs so, we can stitch together any logs that we're making for a given req. Trace id can be used for open telemetry and ... to stitch
other types of info all together.

**Q:** Is trace id for business or foundation? In other words, is it specific to this project and the business problems that we're solving here,
or is it more foundational(can be used even outside of this project)?

Note: Packages in the foundation folder can eventually go into their own repo and become a project.

**A:** It's foundational. We can put it into the web framework of foundation(web folder).

**Tip:** Logging is for maintain, debug and trace the app. Since it's debugging info and it doesn't affect the system while it's running(whether
the logs exist or not, doesn't affect the system), we can put it into the ctx. So if it's there, great, if it's not there, nothing should break and
we can even use a 0 traceID in those cases and it's OK. Everytime we log, we should log with the traceID, so we know what req it's for.

**So put things in the ctx that if they aren't there in some cases, nothing would break.**

We want to store the traceID in the context.

**Tip:** Anytime you're gonna store sth in the context, you need a key and that key needs to be of a **unique type** . Use type aliases to create unique types.
The reason we wanna a unique type for the key is so nobody could just overwrite values that you're storing in the ctx throughout the entire callchain.
A unique type is like `type ctxKey int`. Note that we made it an unexported type.
By keeping the type of the key, unexported, we create an API.

Note: Getters and setters are not APIs! Don't do it! But in the case of context package and the functions there, we get and set values out of a context,
it's sorta a natural thing with the ctx.

Q: Is passing data around in JSON, is that business or foundation?

A: It can be both. If we say it belongs to foundation, it means all of our services talk JSON. If we say it belongs to business,
it means this is a decision that this current project has decided.
The question comes back to: How strict of a policy do you wanna make this? Foundational is pretty strict.

We decided foundational.

### Error handling
**Errors are just values. So they can be anything you need them to be, with any behavior that you need in order to handle the errors.**

We use the error **interface** as the return type. Whoa! Didn't we say we never use an interface as a return type? We did, except when it comes
to error handling. It's actually critically important that we use the interface for the error handling. If you use a concrete type
as the return type for your errors, you're setting yourself up for massive cascading breakages.
So by decoupling errors, we get this flexibility to grow the error handling that we're doing without anything breaking behind it.

What does `if err != nil` mean?

A: In the current implementation of the compiler(and this hasn't changed since go v1.0), interface values represent a two word data structure.
This is an **internal** DS, so it could change. The DS has two pointers(interface has two pointers). One pointer points to some other DS, which
gives us info about the **type** of value that is stored inside the interface and the other pointer points to the actual concrete **value**.

So essentially, an interface value can be in one of two states:
- it could be in a zero value state where both pointers are nil <nil, nil>
- or it could be in a non-zero value state <type, concrete value>

When we say `if err != nil`, we're asking this question: is there a concrete value stored inside the err value(which is an interface value)?
If there is no value, we have a zero value error interface(both type and concrete value are nil), but if there is a value, we know 
errors are just values, so we have an error(second state).

Mental model: Channels in go are for horizontal signaling. Meaning one goroutine signals another goroutine using channels. We think about errors
in go as signals as well. But as vertical signaling. What does vertical mean? 
Let's say we have a func in app layer and it's calling some func at the business layer and that func in business layer is calling a func
at the foundation layer and that func is calling a func from the STD lib and every one of these funcs uses the error interface as a return type and
when we see this, we can think that the error interface is essentially creating a signaling mechanism in itself. Because this error interface
will allow any value to flow through it as long as it implements the error interface.

**Q:** What does it mean to handle an error?

**A:** Error handling rules:
1. you have to log the error. If you can't log it, you can't handle it, some parent layer should handle it. If you can log it,
you have to decide if it's the best place to handle it? If not, you can wrap the error with more context and return it up to the caller.
2. you have to make a decision on whether or not the application can keep running(whether the goroutine that you're on has to terminate or the
entire application has to terminate). You have to determine can I recover from this error and keep going? or sth has to stop? Whether it's the
goroutine or the entire app?
3. because you're handling the error, the error stops with you. It never propagates back up the call stack. If you propagated back up, then it means
you didn't handle it. Note that you can return a **different** error, but the previous error has been handled

Now when the error was returned from the STD lib to foundation library, can the function in the foundation handle the error?
No. Because there is no logging in the foundation layer(our strong policy). So the func in there can't handle it. So what could it then do?
Not much more than wrapping some context around the error. So it wraps some context around the error and returns the error back to the
caller which is in the business layer.

**Tip:** We use fmt.Errorf() with `%w` verb for wrapping the error with more context and errors.Is and errors.As for inspecting the error for anything
that's inside of it.

In business layer: is there a concrete value stored in the error interface? Yes. Can we handle the error here? **Yes because we can log it.**
But we have to decide: Are we the best place to handle it?

**The lower in the call stack that you can handle an error, the better chance you have in recovering and keep going.** The higher that error ends up in
the call stack, the more chance that goroutine at the very least is terminating.

So the business layer(even though it **could** handle the error), decides it's not gonna handle it, so it wraps it with some context and return it to
the caller(app layer), in other words, propagate it.

App layer says is there a concrete value inside the error interface? Yes and it can handle it and it does. How?
We log it. In case of app layer, it's probably the error middleware. It sends a res back to the client and we terminate that goroutine and let it die,
error doesn't propagate anymore. We handled it.

We see error handling in go as vertical-level signaling. We can define any concrete value we want at anytime and use it as a signal.

We don't want the handlers to handle the errors, we want them to return the error back to the error middleware and the error middleware func inspect
what's going on and handle it. We put the error handling into **one place**.
![](img/12-1.png)

Q: What types of error signals or error values we need out of the box?

A: Two values to start with.
- **trusted error type(`RequestError`)**: We don't want to leak anything about the system by default. We have an error that has a message(string) and we
trust the message and that message can go all the way out. **Everything is not trusted to begin with.** If we receive an error
and we don't know what that error is, we don't know what type that error is, we treat it as a 500. But overtime we can improve and make it 
more specific, maybe a 4xx.
- **shutdown error type**: if your service is having integrity issues, it should not be running. When having these issues, the software
should shut down before starting corrupting DBs and filesystems. Note that we don't want to shut the software down forcefully, we want to
have a controlled shutdown. One of the reasons we have a shutdown channel in the web app(main.go), was to give code at any level a chance to suggest
that we should shut down the app. Note that code in the app layer could shut down the app itself, but the layers below the app layer(business and foundation),
are not allowed to shut down the app itself. Those layers are not allowed to call panic(), os.Exit or ... . However, they can **suggest**
shutting down the app by returning a shutdown error. They say to the error middleware: I think you should shut down this app and then the error middleware
can do some checks as well. So they suggest shutting down, but they shouldn't actually do it.
- (there is also an auth error type)

## 13-Error Handling
Q: Is shutdown error type foundational or business?

A: No right or wrong answer. It's how strict of a policy you wanna use. Foundational. It's in shutdown.go .

**In go we use the word `Error` at the end of the name of a type to signify that it's an error type. Like shutdownError.**
Note that `ErrorResponse` is not an error type because it doesn't with the word `Error`. But RequestError is an error type and it's a trusted
error.
It's good to make the error type unexported, so nobody can do type assertions and get out of the decoupling. So when it's possible,
make that custom error type unexported.

The factory functions for a custom error type should still the `error` interface not the custom type itself! Like `NewShutdownError`.

The ordering of things in a package(look at shutdown.go):
- the concrete type 
- factory function of that concrete type
- method set of that concrete type(if they're a lot, you can make them into separate files but in the same package obviously)
- regular functions

**tip:** for custom error types(struct) use pointer receiver and for defining methods for slices, use value receiver.

Q: Is trusted error type foundational or business?

A: Business. It should be versioned as well. So put it under v1 folder because for v2 we might decide we wanna different fields and ... .

We can convert a layer into a package as well. For example the v1 folder(layer) could be converted into a package by adding a <package name>.go
file there and define the package. **Either you have a layer and there are packages inside of that or you have just packages.**
But this situation(making v1 a package and we also have packages inside of v1 package), breaks one of the rules where we have sub-packages so v1
should be a layer but now we just made v1 a package. But we don't have a better answer.

Note: The `mid` package has the middlewares. It's kinda containment which is not good.

We like to use singular names for files.

Note: It's not good to use an alias for an import.

Note: An error returned from the error handler(error middleware) could be an error returned from web.Respond() or could be a shutdown error. But we need to
validate that it was a shutdown error. We do this using the `validateError` func. If the error was syscall.EPIPE or syscall.ECONNRESET,
we don't care, it's not enough to shut down the system. But if it's anything else, return true and therefore we would shut down.

### Panics
We need to stop the panics. The http package already stops panics. So the goroutine that the http package creates to handle the req, will stop the panic
if it occurs and it will return a 500. It's important that we stop the panics for ourselves and allow the rest of the middleware to run.

If the handler panics, the panic middleware should be as close to handler as possible and then the rest of the middlewares go to above layers.
So the panic middleware stops the panic and let the above middlewares to run(like the error middleware).

`recover()` doesn't work unless put in a `defer` statement.

There are two things important with the panics:
1. we should log the stacktrace. You don't want to lose the info in the stack trace in case we stopped a panic.

When a deferred func is running, we're no longer in the func that that defer func(){} was defined and we're also not in the parent call stack.
We're in sorta a matrix.

For example when the `defer func(){}` of the panics middleware runs, we're not in that middleware and also not in the parent middleware!
This is a problem when it comes to the errors because the handler **can** return an error to the errors middleware but the defer func() {} in the
panics middleware(which is between handler and the errors middleware) can't return an error back. Because it's out of that call path. What we need to do
is to construct an error value in deferred func and return it to the errors middleware, but how we're able to get a value back to the parent func
through the `return` statement?

We can name our return arguments. This in itself is not a problem. The naked return statements coming with this, is a problem. Because the return
values are named, with the go syntax, we don't have to specify the values after the `return` statement. This is bad in terms of readability.

**Naked return is bad.** If you **have to** name the return arguments, still don't use naked return, specify the values by using the literal values
like 0 in case of int if that's not what we care about, or use the variable name. So be explicit on the `return`.
```go
package main

func Sth() (myint int) {
	// DO: be explicit on the return
	// DON'T: don't use naked return
	return myint // or return <constant - if we don't care about this>
}
```

In the panics middleware, we're using a named return arg and it's name is `err`. So even though we're inside a deferred func, thanks to closure,
we still have the err var accessible to me and we can still set it right before going to the parent func.

So two reasons for using named return args:
1. for the defer func scenario
2. when a func is returning 3 or more values and all of them have the same type. But don't use naked `return`s.

So now when we have a panic inside the handler, the middlewares will still run, we still have the `request completed` log and
specifically the error middleware will return a 500.

## 14-Metrics / Panics / JWT
The metrics package in business layer is a convenience package.

We have a singleton in metrics package. But didn't we say not to use singletons?

Yes, but because it's almost impossible not to in this case. The expvar package kinda already acts as a singleton-based package. It exposes
those endpoints and behind the scenes, has access to these variables that we're declaring through the package. 
The singleton `m` var is unexported though.

**If you check all these three, you can get away with package-level variables:**
1. the order that they get initialized in, doesn't matter. As long as they get initialized before the main func and there's no ordering at all(you don't
need one package to come before another), you can get away with it
2. you don't need anything from the configuration system. The standalone construction always works. There's no configuration involved. The moment you need
configuration, you can't use pacakge-level var, you have to construct it somewhere(most likely in main) and then pass it around your program.
3. most important one: the only code that should touch that package-level var, is the same file that it's declared in, **not** other files(even inside the same
package)

An example of these rules being followed, is the `m` package-var in the `metrics` package.

Just like cfg, do not scatter metrics all over the program, define all the metrics in one place in the metrics package.

Just like the trace, we define a context key for metrics. Since it's in a context, if it's not there, means we're not gathering metrics therefore
it shouldn't hurt or break anything.

Some metrics are being set in the metrics middleware and one of them is being set in the panics middleware and we could set others in other places of code
as long as you have the ctx.

We're not constructing the metrics type in the main and passing it around. We could, but since we don't need any initialization ordering,
we don't need anything from cfg and the metrics var(`m`) is not used in multiple files, so it's a package-level var and it's also not exported.

### authentication and authorization
- What JWT can do? What it can't?
- What is open policy agent?

We need both authentication and authorization. We need in some cases APIs that you can on;y access if you're authenticated and then there are
APIs that is just not enough to be authenticated, you have to also be authorized, to use that API. Sometimes you're authorized to hit that endpoint
but you're only authorized to see certain data coming out of that endpoint.

We can do all of these with JWTs! **JWTs will give us the ability to do both authentication and authorization.** However, there are kinda some holes
in JWTs. JWTs is a **foundational** tech that most likely you're gonna have to build sth on top of. Out of the box, JWTs isn't gonna be enough for
most apps.

You never wanna roll your own security system!

What does it mean to be authenticated?

Authentication means I know who you are. I've given you sth and you're giving it back to me and I know that that is sth that I've had to give it to you
and therefore I feel confident, I know who you are.

The only way you could've got that token is that **I** produced it and gave it to you.

Authorization: what you're allowed to do. We start with you're not allowed to do anything and then authorization says what you're allowed to do, those
things are specified.

- header: this section of jwt gives us info about how to process the data related to the token. Helps us with the authentication piece
- payload: we call this the claims. Payload helps us leveraging authorization piece
- signature

There are 2 algos to sign data:
- HS algo: creates a signature using a shared secret. You can see this algo on jwt.io .
- RSA algos: is what we're gonna use and this is one using a private, public key-pair.

### RSA
We take the private key and using an API related to JWTs, we can sign jwts.
we take header and payload and then we sign it and the signature gets attached to the header and payload. The string formed by these three is what we
give the client. Note that this data is not encrypted. Because we can take the JWT and decode it! Like on jwt.io . This data is just base64 **encoded**.
There is nothing encrypted. **Do not put private info in the payload of the JWT.** People can read that. We signed the jwt with the private key.
When a req comes in(requiring authentication), we need to authenticate the token. We need to make sure that we're the one that issued that token.
How do we do that? We run the token through an algo that we're gonna get for free with the public key and that algo tells us whether or not this token
was generated with our private key. It's gonna look at the signature of the token and says: yes, your private key generated that signature.
So if our private key has generated that signature of the jwt, you're now authenticated. Anybody can have my public key. Anybody on the planet can
validate that this token was signed by us. We don't want anyone have access to our private key.

**Token is not encrypted, just encoded.**

Key rotation: We found out that someone got our private key and now we got to rotate these keys.

**The authentication happens by taking the signature part of the jwt, running it through a func that we get with our public key and validating it that it had
to be signed by our private key.**

Now that we've authenticated the jwt(which means it was signed by our private key), we can now look at the payload data(claims). This was for RSA algos.
For HS algo, we just validate that that signature was signed with the shared key.

A middleware can say: anybody hitting this route must be an admin and ... . This is authorization.

Q: When do you need to potentially build more on top of these?

A: For example, we've given someone a jwt that has a claim field that is for admins. But let's say we gave someone a token with admin field set to
true in the claims but with expiry of 100years. Now let's say we don't want this guy to be able to login anymore. How we're gonna do that? He has a token
with long-lasting expiry. Well we can do a key-rotation and say: Any token that was signed by this old private key is no longer valid, but we've
got 100,000 of users and now all of them have to get a new token because we didn't want that specific user access the system anymore.

So just using the bare-bones jwt implementation ends up not being enough.

In a real world system, yeah we give the basic jwt to the user, but also you also would have a DB call to the user table. Now in case of previous example,
we see in the DB that that user is blocked or sth and now when his req with the token comes in, he's authenticated at the token level but then when
we make the DB call for the next step of authentication, we see he's blocked or sth.

So you can't rely on the jwt itself alone to do the authentication. You want an extra check like a DB call for stopping someone from accessing the system
without invalidating everybody's tokens.

So we would have a second level of authentication check and that's why maybe the client's userId is in the claims. We put some info(not secret info
because jwt can be read by anyone) in the claims so that it can be used on the second level of auth. We can use the **sub**(subject) which
usually is the id of the client that this jwt was issued for.

If you use OPA in the right way, we can get rid of any go code that does authorization. You push the processing of authorization to OPA.

There are two ways of using OPA:
1. sending req to the OPA server(completely offload that)
2. or doing it in proc but detach it by executing OPA scripts directly

## 15-Authentication / Authorization
### Creating private and public keys
#### using CLI
```shell
which openssl # on mac, it's installed by default at /usr/bin/openssl
```

pem is the encoding we use to write the keys(private and public) to disk.

```shell
# generate private key
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
```

Once you have a private key, then you can take that private key and generate the public key pair:
```shell
openssl rsa -pubout -in private.pem -out public.pem
```
Hide the private key in sth like vault, but it's no big deal if anybody has the public key. We need the public key to validate the signature of JWTs.

There is also one check that we can do(we don't need to), but it's good practice that when you're verifying the token's signature, we also verify
that it was issued by you.

The private key we're gonna use in our project zarf/keys/<filename>.pem . Since it's shared in repo, do not use this private key for anything!

#### using Go APIs
For generating private and public keys, look at the `scratch` package in the tooling layer and it's `genKey` func. The scratch package is a 
scratch program for playing with authentication and authorization. This package has basic program for:
- generating private and public keys using cli and go APIs
- jwt generation
- jwt validation at the go level and OPA level
- authorization validation at the OPA level

For generating JWT, look at `genToken`.

If you paste the generated JWT, in jwt.io it says: `Invalid signature` because we haven't given the debugger(jwt.io) the public key in order to
validate the signature(that's the authentication piece - we signed jwt with the private key, now we wanna make sure the
public key validates the jwt signature). So copy the public key and put it in the `verify signature` block in it's first box(public key box).
Now it should say: `signature verified`.

Note: jwt.io is the jwt debugger!

Now to validate the jwt signature that the client sent to us, we can create a parser using jwt.NewParser() .

We wanna use OPA for both authentication and authorization. We wanna offload that out of the Go side. So instead of using the APIs
of jwt package to validate the signature of jwt, we want OPA to do that.

In the `rego` folder, we have rego scripts. By hard coding the rego scripts, we have bounded to the binary to run those binaries. Ideally,
we would like to pull those scripts from a server or even better, have an OPA server where we call it and those scripts live in the OPA server.

Instead of using the go-jwt library to do the jwt signature validation, we wanna use OPA.

## 16-Authentication / Authorization
Q: Why do we put the auth package in the business layer? Couldn't this be foundational?

A: Because the `Claims` struct is business-related(project-oriented). But couldn't we just put the bulk of the package in the foundation layer
and then put that piece in the business? We could, but this package is a convenience package, there's not a lot of code to break up.
You might want to break this package into a package inside foundation layer and a package inside business layer.

**Tip:** At the business layer, we don't want to do validation, we want validation to happen at the application layer.

### Enum pattern
Anytime a value could be an enumeration, we apply this pattern. Anytime I have a value like a string that's being passed in as
external input that needs to be validated, I follow this pattern. For example user roles follow this pattern. We want the compiler to help us here.

Go doesn't have enumerations. Why? Because of engineering tradeoffs. One of the reasons is because constants in go are not read-only variables.
Constants in go have their own type system. A constant in go is literally a hard-coded value that's laid into code during compilation. They're not variables.

Since constants have their own type system, numeric constants have this characteristic:

The minimum precision of a numeric constant is 256 bits. Your go compiler when it comes to constants, is a high-precision calculator, but when
it comes to integers(variables) it's silly because every constant eventually gonna be assigned to a variable in our code and you can't fit 256 bits integer
into 64 bits.

Because of these tradeoffs(which is a constant can be implicitly converted at compile time, to variable) you can't have enumerations.

Define a type but use an unexported field, normally we call it `name`. An example is the Role type in the user package. That type represents an enum.
Then we can define an unexported map of all possible values of the enum(roles map for example).

We call this a set instead of enumeration, like a set of roles.

We can't construct a set(enum) by hand, we should use the exported ones by the package. Q: I can construct a zero-value set. Right? Yes, but it's
not valid.

**Tip:** The parse(like `ParseRole`) and Must(like `MustParseRole`) funcs are idioms that exist in STD LIB. The Must is a func that if a bad thing happens, it
would panic instead of returning an error. Using the Must funcs in test is good, but you shouldn't be using them in app code.

Q: What is the kid in KeyLookup interface?

A: It's key id. On the tokens(JWTs) that we're going to create, we will add this kid to the header of jwt.
The idea of key id is for key rotation support. You might be using a third party system like vault and what that's doing is it stores an id
for your public keys. So each one of your public keys are assigned an id. Now we wanna know which
public token should be used for a jwt that hits our system. For this, as we generate a jwt, in the kid field of the header of jwt, we're gonna
put that id of the public key pair. So when the token comes back, we look at the kid field in the header of token and that tells us which public key
would validate that jwt and if we don't find the public key in the key store, maybe we've rotated it out, so the jwt is not valid anymore, we don't
even have the public key anymore to validate the jwt.

So we would add the kid to the jwt header.

**Note:** Remember none of the packages other than the main.go file, can access the config system directly. The config has to be passed in. We can
define another `Config` type for a package which has all the fields necessary for the construction of that type, like the `Config` type in auth package
and that type should be passed as value not reference because it represents pure data, doesn't represent an API.

The way vault is used in this project, is good for a dev env, there's probably better ways to do it in a k8s env in prod.
We're not gonna use it in this project, instead we implement our own keystore but you're not allowed to use it in prod. But you can use this package
in testing.

## 17-Auth / Liveness and Readiness
We have created a private key to the project(and put it in repo which is not secure at all, don't do this) under keys folder and it's name is a uuid.
Now that filename is gonna be the `kid` or key id which is put into the jwt header.

We need the keys to be part of the docker image for all this to work(I think it's a security concern).

```shell
make run-scratch # to create jwt

export TOKEN=<jwt from previous command>
make test-endpoint-auth
```

After implementing authentication and authorization, we can almost start focusing on the business problems which is building a REST API for our sales.
To do that, we need to think about our data model.

We're gonna start at the business package instead of thinking about tables and relations at the database level. But traditionally, we start
at the database level.

We don't want our liveness and readiness handlers to be on the same port as the application traffic. Because I don't need that to go through
the middlewares of the app like metrics or ... . I want this on the debug side.

**UPDATE: Bill changed his mind and moved the readiness and liveness handlers to the same port that our api(app) is running.**

We put the readiness and liveness handlers in the `checkgrp` package inside

**Readiness:** I'm ready to receive traffic. I'm alive, but I'm more than just alive. Like if you send me a req, I should be able to process it with no problem.
Also I can talk to the DB.

**Liveness:** Just a ping. Are you alive? We can use the liveness handler as a way of me being able to also get info out of the service, like
the build version, pod IP, GOMAXPROCS and ... . The ingress or other k8s objects don't care about this info that we return in
the liveness handler, it's just for us.

We're picking one req every 100 reqs to set the number of goroutines for metrics. Because we don't want to build the timer loop.
Note that the liveness probe is a free timer loop. Because k8s is calling it on whatever interval you want the k8s to call it on.
So why would you ever need the timer loop? So we can update the goroutine count in the liveness probe handler.

Q: Shouldn't the readinessProbe and livenessProbe be in the dev/sales instead of base/sales? Maybe you want different values for different environments?

A: If you need different values for different environments, that should be in the patch.

The ingress also hits liveness in addition to other components of the cluster. So you might see the liveness handler gets hit twice, so one is from
inside k8s and one from ingress.

Liveness should hit before readiness.

```shell
make liveness-local # to hit the liveness handler manually
```
With this, anytime we want, we can get any info we want from the service and it's also called by k8s on an interval.

## 18-Business Package Design
**IMO it's risky to run liveness and readiness probes on a different port than the API(app) port. Because if the app doesn't come up,
we won't catch it until it reaches the user traffic. Yes it's correct. I just hate the idea that those probes go through all the middlewares and ... .
So we have to hard code in the middlewares that if the req is for liveness and readiness, leave it alone.
It's better to put the liveness and readiness handlers in the same port as our api is running instead of the debug port. So move them into a handler
group called checkgrp(look at v5 project).**

### DDD
We have the handlers in the app layer, it's job is to receive external input(API).

We want the domains to be self-contained. So a user should have pure data related to user and nothing else. Now if in an endpoint, we want to return
products and their related users, this is an application layer problem, not a business layer(where all the domains exist and they're pure) problem.

We're gonna have application-level model and business-level model. So when the business-level model changes, it won't break the
application-level model(the model that is sent to the client as the response).

Note: You can not join any data between domains, even behind the scenes.

If we want the domains to be truly isolated(firewalls), then we need to plug in a storage layer to each domain. The storage layer is going to
be given an API by the domain and the job of storage layer is to store that data however it wants in some DB. From a business layer perspective,
the way you wanna think, is that each domain is running it's own DB, it's own instance somewhere on the planet, but we don't know where. It means
that if we represent each one of the domains as a table in the same DB, then since they're in the same DB, we can do JOINs, right?
No you can't! We said in the DDD world, we need to pretend that the data we're storing for domain 1, is in a physically different DB than domain 2.
So if we want to blend(mix) any data together, I can't do it at the storage layer, we can only do it at the app layer right now.

So again: Even if the tables of our domains are in the same DB(which is the case in day 1), we're pretending that it's not.
**So we shouldn't have any SQL in the business layer(it's storage) that is JOINing two domains.** It is ok to have **one** domain of multiple
tables where we JOIN them. Essentially their storage is together. For example we might have a couple of tables for one domain like user.

Now when a domain gets too big, we can move it from our service and put it into it's own service. Nothing would break because the domains are isolated,
the domains are not dependent on other domains. Now after doing this, the only thing that has to change is the app layer. How it gets the data
from the removed domain? Well the moved domain is technically still in our old service but it's storage layer has changed in a way that
for doing the CRUD, we hit the new service! So it's storage layer is gonna call the new service.

Since product has a userID in it's model, it can import the user package, because the relationship is there.

Now if we want the products + username, we could create a new domain named `productsUsers` which we call it a **view**. This domain also has it's own
storage layer like other domains which also means we assume it has it's own DB. With this, we kinda have a duplication of data of both users domain
and products domain, but more importantly, how does the data get to the DB of this view domain? This is where some of the ideas around event-driven
systems comes into the game. Maybe every time we add a product, that domain fires and event that says: this product created. Now the `productsUsers`
domain subscribes to that event and say: Everytime a new product is created, I want the event and when that event occurs, it gets
the product model back and since the product model has the userID, it can call the user domain to get the user

The fact that domain A has relation to domain B, it means it can import domain B APIs and package. But not the other way around.
For example the `productsUsers` domain has a userID in it's domain, so it can import the user package but not the other way around to avoid cyclic imports.

Now if you don't have time, we can use a single DB instance where each one of the domains has it's own table. But each domain is pure.
Now there's a cheat. With this cheat, if you wanna pull a domain out into it's own service, we would have some refactoring.
The cheat is: Instead of adding the event system just yet, we can create a DB view for our view domains. **So every view domain gets it's own
DB view where we can use JOINs(because it's a domain that contains multiple things)**. Now the view domains don't have to call into other domains.
With this, later on, when I have more time, we can wire in the event system and stop using the DB view for the view domain and start putting it
into it's own table.

DB view is a cheat that we can use to postpone the use of event system. When we're cheating, we still keep the firewalls in the domains
but there's no firewall in the DB layer which should be(separate DB for each domain). But after we have some time, we can
use ddd better and remove the cheat which means turning the DB view into a table and use the event system.

If you keep at least the core business packages(domains) firewalled,  cheating at the storage layer could potentially never hurt you.

When you're at a point where every single domain is managing it's own data independent of the other, even if it's in the same DB, and you work
with the other domain's APIs, you can scale horizontally and also you can pull sth out and scale it vertically if you need to. So we don't want to see
a JOIN outside of a view. We only want to see it in a view domain and even better in there, we'd like to get rid of the DB view and use the event
system to simulate that DB view. With this, we can pull out the view domain and put it into it's own service.

Note: The events don't have to be asynchronous, they could be synchronous.

## 19-Business Package Implementation
Do not create a types package that holds the types for all the domains. Each domain has it's own types.

We want to use custom types  in the models of the business layer. Like the mail.Address type for email instead of string. Because otherwise
someone could hack to abuse the parse func to validate their email.

Bad philosophy: Every layer(app, business, storage) should not trust the layer above. So we validate the data before working with it in business layer
and storage layer. Yes when app layer code is talking to business layer, we should validate the data at the business(no trust).

But for example when there's a relationship between two packages, for example when we have business to business API calls happening like
product and user, product is importing user, but the data is already been validated 99 out of 100 times. So we're validating unnecessary and it could
include DB hits.

So this idea of no-trust, is creating massive inefficiencies when it comes to business-to-business API calls. Maybe we could put a flag in the API
and say: hey, this is a business to business call. No can't do that, it's not good.

Instead, we should create layers of trust and move validation somewhere else.

- We know only the business layer calls the storage layer. Tge APIs between business layer and storage layer are trusted. So storage trusts the called
with all the data. 
- We also trust the business-to-business calls. Note: There are times where you have specific business rules, that's different(you can
do validation there). It's business logic not data validation, it won't create inefficiencies like having to call DB multiple times for the same validation.
Here, we're talking about the actual input of data coming into the API. By doing this, it pushes validation at the application layer.

Now by having a type system at each business domain and also with those parse funcs, we can mitigate blindly trusting. For example for creating
a user, we should provide a `mail.Address` not a string and the only way to create a mail.Address is through mail.ParseAddress(<string>) .
Yeah it could create it with a mail.Address{} as well. The same for `Role`. WIth this, the business is not worried that the app layer wouldn't validate
the data, because without using those parse funcs in the app layer, it can't call the business layer.

Note: Zero values can escape this idea.

We want to have different models for different CRUD ops. Like NewUser, UpdateUser. Otherwise if we used `User` for new op, the caller
would think it has to give us an ID because there is an ID field in the user but not in NewUser.

About update: Somtimes we can't update a single field, because some fields can have certain values if other fields are certain values and again if other
fields are certain values and ... and so you can't just update individual fields, you have to update an entire model. But if I wanna update sth,
I shouldn't be asking the caller to give me a complete model necessarily. That's a burden. So we use pointer semantics in the update model as a way
of describing the concept of null.

Nil vs null: NUll is **absence of value**. Now go doesn't natively have the concept of null. It has the concept of nil. nil in go represents the **zero
value** of types that can work with pointer semantics(pointer variables and all the internal types that are pointer).

We need the concept of null(absence) in the update model. So we use pointer semantics, so user gives us only the fields he wants to update.

Note: Remember that the data model of user package is `user.User`. **A lot of times, one type of a package would have the package's name like the
`Context` type in the `context` package.**

Now in the user package, the data model is user.User, but we also have an API that we need and for the API we use the name `Core`. We have a type and
a constructor for it. Why core? Because remember we talk about "the core business package", "the core user package". So we named it Core.

An API defines the type system it needs to accept data and to return data. The package itself decides whether or not it needs **decoupling** or not.
For example the business core user package needs to be decoupled from the storage implementation. So it's defining
the Storer interface(not the other way around). Because it's saying: This is what I need in terms of behavior for my API to work.

**Therefore: If a package needs decoupling, which means it needs an interface, it itself must define that interface not the user of the package.**
Like the relationship of a service and a repository. The service needs to be decoupled from the storage layer(repo), so it defines the repo interface
inside itself(service defines the repo interface inside the service package).

Whenever a func is IO-driven, so it should take a ctx for cancellation and you might want to trace it.

Note: When defining a method on a type, if that type represents an API, the method receiver should be pointer, otherwise use value receiver(also
if we mutate the data of that type, use pointer receiver). For example `Core` type of user package represents an API not data, so we use
pointer receiver for it's methods.

Note: We use value semantic instead of pointer semantic for second arg(NewUser) of Create in user package, because that arg represents a data.
Also the return type of that func is a value, not a pointer, because it represents a data.

Interface types -> value semantics

Note: Decoupling is handled through interfaces.

**Note about API design of update method in user package:**

Why we want both User and UpdateUser model?
Remember the Update method is about asking the caller what field they wanna update? You need to get a snapshot of a User model and
update that one based on UpdateUser. So we need both. But the question is: Can't this method pull the User out of storage itself?
So the called won't have to pass it?
We could. But the reality is that the caller probably already had to get the User for some reason before calling this method. Especially if it's a
business-to-business call. They probably already have the user.

With this API design, we're actually forcing the app layer to go get the User first(by getting it from client or maybe call into business layer
to get the User) before they can do an update using the business package. There's efficiency in that, because if there's a business-to-business calls
being done, you(caller) probably already got a User so we don't need to query it again in the Update() method! We just pass it.
So don't change the API to accept a userID instead of whole `User` model. If we accepted the userID, it would be inefficient because
in the Update method we would have to query for the User model. We would already have the user in the caller.
We don't want to do that in the business layer, because that's gonna be done potentially a lot of times before calling this method. We don't want
to get small values(and not whole model) and again do another query in the business method.

The same thing for the `Delete` method. It gets the whole User model. By designing the API like this, the caller(app layer or another business) is **forced** to first
find out if the user exists to delete it(since it needs to pass it to the Delete method).

The methods that are put into an interface are APIs.

**Who does validation? The app layer. It means the business layer does what the app layer wants and the storage layer does what the business layer
wants.**

Any query API that's gonna return a slice, should have filtering, ordering and pagination. 

Q: Hey the Query method gets 5 args, isn't it better to use a struct?

A: No using struct would make this API worse. A struct hides too much, it's not precise as individual params.

Q: Why QueryByIDs doesn't have filtering, ordering and more importantly pagination? We could potentially return 1 million records.

A: Yes we should!

## 20-Business Package Implementation
### Strategies about filtering and ordering
The validation tags(`validate`) are gonna be in the app layer, we don't want them in the business layer. It's the app layer job is to do validation.

You can have package-level variables when you don't need:
- configuration
- instantiation of them 
- order of their construction

For example in validate package, we don't have construction using New... funcs. If we do care about 
different languages(currently we only support En), we would have to construct it in main and pass it around, instead of having package-level vars.

Note: Avoid singletons.

By applying validation, it requires us to have another error type(errors are sort of signal) named `FieldError`. Because if any of those fields
fail validation, you wanna report that.

Note: The `RequestError` of service v4.1 project is the same as `TrustedError` in service v5, just with a different name.

We used value receiver for methods of FieldErrors because you could have more than one field fail validation and you want them all.

Since we added another type of error, we need to handle it in the errors middleware(add a new `case`).

Do not add validation tags when you're relying on the type system for validation in the business layer. For example in the `QueryFilter` structs,
we don't add validate tags for the fields that have a specific type so that we can rely on the type system for validation, for example using
`mail.Address` or `uuid.UUID`.

Note: Product core imports user core. We can do this because a `Product` has a `UserID` field. So we can bring in the `user.Core` and access
the user domain in product, when we need to. But the user package cannot bring in the product domain if it even needs it. The relationship isn't there,
a user model does not have a productID. But product does have a userID, so we can bring in the user package into the product package. This is when
one core can use another core. So we won't get any import cycles.
But if in the data model, a user has a productID and a product has a userID, the user model is wrong.

`cview`(core view) is the layer that we put the core business packages that are view-based. We're gonna have long names in there because the views are merging
data together. We don't have just a user like what we have in the core package, we have sth like user summary data. The naming convention for this
layer is: <the main domain like user><other info ...>.

Opinion: You don't need the cview layer(folder). The argument is: I don't care if this domain is a view Because the packages that 
we're putting under cview, are core packages like any other core package.
This is a correct argument. The fact that this domain(domains inside cview layer) are hitting our views, are irrelevant. They're core.
But for our mental model, we separate core packages and view core(cview) packages. But note that it has the same files as core packages, files like
model.go, <domain>.go, order.go, filter.go and ... .

cview domains have it's own domain and table(a DB view). If there are a lot of reports going on, we would have a lot of cview domains rather than
core domains. We don't want to clutter our core layer. So when we get into reports, we wanna use cview packages.
Views exist because we have to join data across existing domains. But there's still a primary domain in that join.

Note: We have layers and packages. Layers are just folders, but packages are folders with at least one go file in them.

When it comes to cview packages, you only need the query operations. But eventually if you decide to wire-in events, you would have a `create` and
`update` because when those events fire in other domains, you wanna update this view.

Note: As mentioned, when using DB views to simulate a separate DB, we're cheating.

Instead of using `usersummary` as the package name in cview, use the same folder structure as in the core layer(preserve the core domain),
so create a user package that has a user folder and inside that folder, the views that their main domain is user, exist.

DB views are the only place that we can use JOINs between domains. But eventually at some point, you wanna get rid of the views and use the event system
to make this a table.

## 21-Database Support
Your DB shouldn't be running inside the k8s cluster. Because the cluster is about the ability to be restarted. But there are DBs that have been designed to
run in k8s. But traditional DBs need to run outside of cluster, in a cloud env. But in the local env(dev env), we wanna bring it up and
down quickly, so we run DB inside the cluster. However, we're gonna use a PV, so if DB does get restarted(which definitely would do),
we don't lose the data.

Note: The dev-database.yaml is only for dev env and it's not production-ready.

Install pgcli.

Run `dev-up-local` to pull the postgres image into kind. After running it, we should have a DB pod running.
Now run:
```shell
make pgcli
```
Now you should get the prompt.

### Go support for DB
I don't like ORMs.

sqlx package is an abstraction layer for executing SQL in a safe way(avoid sql injection). This package leverages the go std lib sql pacakge underneath
which requires a DB driver. There's two drivers today:
- pq as the driver(this project is done)
- pgx itself as the driver(actively being worked on). pgx tries to be sqlx too(being abstraction layer in addition to DB driver). So we can use pgx
not just as a driver but also as abstraction layer for using sql. Note: You don't need convenience funcs around sqlx or pgx, you can just use it.

In the service project, there's the ability to either use pq or pgx as the driver(see business/sys/database/). You see two packages there.
pq and pgx packages. They are convenience package(`database`) using sqlx that we've built to access pq and pgx third-party packages easier.

The service project is using the pgx driver.

The dbarray package is code that we actually got from the pq library to make arrays work.

Q: Why database package is in the business layer?

A: Because we're gonna add both logging and tracing in this package and those are not foundational.

`queryString` is used to print the query. But there are times where you can't log those queries since it's sensitive info. So don't use
that func for those queries.

Do not abstract the running logic of the service into a `run` func Put all in the main func and use divider comments. Because then you can leverage
closures and ... .

## 22-Migrations
We use `ardanlabs/darwin` for migrations.

Once you run your migration, you can't go back and edit anything in an old version of migration.

Note: We don't use the seed information for testing. We should re-design the testing to never use the seeds as data for testing.

dbmigrate package is a convenience package around using the darwing module.

The `admin` program in app/tooling is for learning migrations. To run this program:
```shell
make migrate

# then do:
make pgcli 
```

We run the DB migrations through the `initContainers` through k8s. We can do this in dev, staging and production envs.

Make the default configs work in every environment. We can use telepresence to use the same namespace which avoids changing 
things in dev and other environments.

## 23-Storage Packages / Handlers
We have a `stores` layer(folder) in each core domain and we put storage implementations in that folder. Now why a separate layer?
Because we could have multiple implementations of that `Storer` interface.

We use <domain name>db or <domain name>Cache or ... . Notice the <domain name>. We use the <domain name> in order to avoid naming clashes
when importing.

Note: The app layer has it's own models. Also the business layer has it's own models. Also the storage layer has it's own models(for example, the model.go file
in `userdb` folder). We have to do this. We can't share these models across these boundaries. Because each boundary has to define for itself, how data
is moving in and out of that boundary. So we'll be marshaling the models between these layers.
Note: The upper layer that uses the lower layer interface, can also use the lower layer models but not the other way around. Imports always go down not up.

Sometimes these models of different layers are identical but their struct tags are definitely different.

When converting things using funcs, use `to` in naming not `from`. Because it tells you what's going out. Like `toDBUser` or `toCoreUser`.

We want to work with locale time in business layer. This helps with logs and ... . Especially in dev mode. For example if we're developing and
we have a 5 hr difference with UTC and we see the logs in UTC, our brain would explode. Instead, we want to see the logs in our locale time.
This doesn't affect the frontend.

Since DB has it's own type system and different field names, we need separate models in storage layer.

**Do not use struct tags for business layer models.**

The models in app layer, are prefixed with `app`. These models are exported because we wanna use them in tests. There's another approach
where we make them unexported and redefine them in tests again. You can make them unexported like the db package models.

All of the types of models in app layer are built-in types. We don't use UUID or mail.Address . But why?

It's because of shortcoming of how the std lib json unmarshaler funcs work.

Bug in std lib: The json decoder, doesn't return a detailed set of reasons the unmarshalling failed. Because we don't know why the unmarshalling
failed, if we use non-builtin types in our models, all we can return back to the user is we couldn't unmarshal it. We can't send back detailed info.
There are 2 solutions:
1. fix the std library decode to get back detailed info
2. use a different json package that gives us detailed info

Instead, we can use builtin types for our app models, but then validate that for example the string type is a UUID. So that we get back detailed info
about which validation failed. So instead of putting non-builtin types into app req models(like AppNewUser), we use validations on models after
the initial model that deal with reqs. So we use builtin types so that the json decoder doesn't fail and if payload is not correct, the failure happens
when we validate the payload(not when we unmarshal).

So what we do is we use builtin types in order to prevent the unmarshal to fail and then we validate using the types we really want.

Anytime you have a func for parsing(like parseOrder or ...), you should validate stuff as well.

`NewRequestError` creates a trusted error.

## 24-Testing


## 25-Testing / Observability
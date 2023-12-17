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

If you had a package that doesn't make sense to have a file named after the package name, it's a containment package. For example the `mid` package
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

Q: What does it mean to handle an error?

A: Error handling rules:
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
- trusted error type: We don't want to leak anything about the system by default. We have an error that has a message(string) and we
trust the message and that message can go all the way out. **Everything is not trusted to begin with.** If we receive an error
and we don't know what that error is, we don't know what type that error is, we treat it as a 500. But overtime we can improve and make it 
more specific, maybe a 4xx.
- shutdown error type: if your service is having integrity issues, it should not be running. When having these issues, the software
should shutdown before starting corrupting DBs and filesystems. Note that we don't want to shut the software down forcefully, we want to
have a controlled shutdown. One of the reasons we have a shutdown channel in the web app(main.go), was to give code at any level a chance to suggest
that we should shutdown the app. Note that code in the app layer could shutdown the app itself, but the layers below the app layer(business and foundation),
are not allowed to shutdown the app itself. Those layers are not allowed to call panic(), os.Exit or ... . However, they can **suggest**
shutting down the app by returning a shutdown error. They say to the error middleware: I think you should shutdown this app and then the error middleware
can do some checks as well. So they suggest shutting down, but they shouldn't actually do it.

## 13-Error Handling
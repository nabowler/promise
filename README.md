# promise

A simple POC of Promises/Futures in Go.

## Why

¯\\\_(ツ)_/¯

This is the implementation of a thought experiment of what Promises/Futures may look like in Go.

I am unsure, as of this time, when this pattern may provide cleaner or clearer code than
direct use of channels and goroutines, but that is not to say that those times do not exist.

## Usage

```go
import github.com/nabowler/promise

func me(ctx context.Context) {
    // create a Promise that can be fulfilled in the background
    p := promise.Me(ctx, func() (*http.Response, error) {
        return http.Get("http://example.org")
    })

    // other work to do
    // the promise will* run in the background
    // *subject to the whims of the runtime scheduler

    // call the promise to get the result
    // this will block until the promise is resolved
    resp, err := p()

    // handle results
}

func you(ctx context.Context) Promise[*http.Response] {
     // create a Promise that you will fulfill with complete
    p, complete := promise.You[*http.Response](ctx)

    // some longer running process that will eventually complete the promise
    go func() {
        resp, err := http.Get("http://example.org")
        complete(resp, err)
    }

    // yield the promise to the consumer(s)
    return p
}
```

### Guarantees

For the purposes of the following, "Promise" and "Complete" will include both of the types `Promise` and `PromiseNoError` and `Complete` and `CompleteNoError` respectively.

* Calling a `Promise` will block until it is resolved, either by the Complete returning a result or the Context being done.
* Calling a `Promise` multiple times will return the same value(s).
* Calling a `Complete` will never block.
* Calling a `Complete` a second more more times will not affect the return value(s) of the associated `Promise`.
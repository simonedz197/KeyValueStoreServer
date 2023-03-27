# Week 2 Reference Implementation

> This reference implementation is to exercise the test harness (think of it as 
> test code for the harness). It uses third party libraries which you shouldn't
> be using (e.g. Gorilla, logrus, etc.), is written in a style that may not 
> match that expected for _your_ solution, (read: it is somewhat thrown 
> together) and cuts a bunch of corners (the worse offenders are commented). 
> While you are welcome to take a look at how this project is put together you 
> should **not** use it as a basis for your solution.

## Usage

```
./store --port <port> [--depth <depth>]
```

With just the `--port` flag the following are exposed:

* `PUT` `/ping`
* `GET` `/login`
* `PUT`, `GET`, `DELETE`, `/store/<key>`
* `GET` `/list`
* `GET` `/list/<key>` 
* `GET` `/shutdown`

Authorisation is dynamic with the bearer token being used if present, otherwise
the authorisation header is considered to be the username.

The implementation always uses LRU, but with the default depth being `MaxInt`, 
effectively disabling it. Setting `--depth` will enable LRU behaviour.

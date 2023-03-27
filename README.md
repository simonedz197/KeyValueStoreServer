# HTTP REST Server

This project takes what you have learned in the past few days and applies it to
an HTTP REST Server which you will write. The main task allows you to
demonstrate your understanding of the core concepts in Go. If time permits you
can also attempt the optional tasks. A full test harness is provided allowing
you to check your solution, however, this should not stop you from writing your
own tests.

## Core Requirements

An HTTP Rest server that will act as an in memory key value store. Users can
`PUT` values into, and `DELETE` values from the store _provided_ no other users
is using that key. Users can `GET` values from the store regardless of who
created the value.

The key/value store _must_ be thread safe.

> Note: This is a very contrived specification designed to exercise certain
> areas of knowledge.

### Working

Although week 1 explored a Key Value Store this project should be started from 
scratch with concepts and ideas being reused. You can of course reuse any code 
from the week 1 labs but that is not a requirement.

While this is an individual project it is assumed you will work on it as you
would a client project. This means you are encouraged to collaborate with
colleagues and ask questions when stuck or unsure of something.

Aim to keep external packages to an absolute minimum. You should only need to
download a JWT and Bcyrpt/Argon2 implementation to complete the optional parts
of the exercise.

### Startup

The store will be started with the command:

```bash
./store --port <port>
```

Where `<port>` will be the port to bind to. Failure to parse the `<port>` should
yield an exit code `-1`. Failure to bind to the `<port>` should yield and exit
code `-2`.

### Logging

The application should log all requests to a logfile (e.g. htaccess.log) 
showing:

* The time of the request
* The source IP of the request
* The HTTP method
* The URL

The application should log all other information to a separate logfile (e.g.
store.log).

### Ping

Respond to a `ping` request with a `pong`. The `/ping` endpoint is a simple way
of ensuring the server is running.

#### Request

```http request
GET /ping
```

#### Response

```http request
200 OK
Content-Type: text/plain; charset=utf-8

pong
```

### Put

Attempt to create _or_ update a `<value>` stored under `<key>`.

```http request
PUT /store/<key>
Authorization: <username>
Content-Type: text/plain; charset=utf-8

<value>
```

If `<key>` does not exist, or `<key>` exists and was created by `<username>`
then store `<value>` under `<key>` and return:

```http request
200 OK
Content-Type: text/plain; charset=utf-8

OK
```

Otherwise, return:

```http request
403 OK
Content-Type: text/plain; charset=utf-8

Forbidden
```

### Get

Retrieve a `<value>` stored under `<key>`.

```http request
GET /store/<key>
Authorization: <username>
```

If the `<key>` exists:

```http request
200 OK
Content-Type: text/plain; charset=utf-8

<value>
```

If the `<key>` does not exist:

```http request
404 Not Found
Content-Type: text/plain; charset=utf-8

404 key not found
```

### Delete

Attempt to delete a `<key>` and its value.

```http request
DELETE /store/<key>
Authorization: <username>
```

If the `<key>` exists and was created by `<username>` return:

```http request
200 OK
Content-Type: text/plain; charset=utf-8

OK
```

If the `<key>` does not exist:

```http request
404 Not Found
Content-Type: text/plain; charset=utf-8

404 key not found
```

Otherwise, return:

```http request
403 Forbidden
Content-Type: text/plain; charset=utf-8

Forbidden
```

### List store

Return information on the store.

```http request
GET /list
Authorization: <username>
```

```http request
200 OK
Content-Type: application/json; charset=utf-8

[
    {
      "key": "<key>",
      "owner": "<owner>"
    },
    ...
]
```

### List key

Return information on a `<key>`.

```http request
GET /list/<key>
Authorization: <username>
```

If the `<key>` exists:

```http request
200 OK
Content-Type: application/json; charset=utf-8

{
  "key": "<key>",
  "owner": "<owner>"
}
```

Otherwise

```http request
404 Not Found
Content-Type: text/plain; charset=utf-8

404 key not found
```

### Shutdown

Shutdown the application gracefully.

```http request
GET /shutdown
Authorization: admin
```

If the authorization is `admin`

```http request
200 OK
Content-Type: text/plain; charset=utf-8

OK
```

And exit with code 0. Otherwise, return:

```http request
403 Forbidden
Content-Type: text/plain; charset=utf-8

Forbidden
```

## Advanced Requirements

The following requirements are stretch goals. While they do not have to be
completed for the course it is strongly advised you at least attempt them to
get an idea of the concepts they are testing.

Each advanced feature provides a _capability_ the test harness can check (See
**Testing** below).

### Login

> Capability: `login`

Allow a user to log into the system. Login will return a Bearer token for
authentication on future requests.

```http request
GET /login
Authorization: Basic cm9vdDpwYXNzcGhyYXNl
```

Successful login will return:

```http request
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8

Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mjc2ODE1MzAsIm1vZGVsIjoiZGVmYXVsdCIsInVzZXIiOiJyb290In0.dV5VxBgo8XuMofAtPtu2Ac60v5ds8B0lNYB9wh5J_tkapQ7AXs1YHSaCdEB-uV4r7a0EBuBBD2pkobtrncuqcQ
```

Failure to login will return:

```http request
401 Unauthorized
Content-Type: text/plain; charset=utf-8

Unauthorized
```

All requests that take an authorisation header will now expect the bearer token
and the username will be retrieved from that. Invalid bearer tokens will
result in:

```http request
401 Unauthorized
Content-Type: text/plain; charset=utf-8

Unauthorized
```

Missing bearer tokens will result in:

```http request
403 Forbidden
Content-Type: text/plain; charset=utf-8

Forbidden
```

`/shutdown` will only respond to the admin user.

The server should allow for the following users:

| Username | Password  |
| -------- | --------- |
| user_a   | passwordA |
| user_b   | passwordB |
| user_c   | passwordC |
| admin    | Password1 |

> Consideration: User management is not part of the spec for this application
> but it is good practice to use something like bcrypt or argon2 to hash the
> passwords. Even though the user list will be hard coded see if you can use
> hashes for the passwords rather than storing the plain text version.

> Stretch goal: Login endpoints can be used to enumerate the users. Can you
> make login constant time to avoid this?

### Admin Override

> Capability: `override`

Allow the admin user to `PUT` and `DELETE` keys they do not own. `PUT` will not
change the ownership if the key already exists, it will simply overwrite the
value.

### LRU Store

> Capability: `lru`

Constrain the size of the store and make it Least Recently Used (LRU) for
key eviction.

Startup will have a new flag:

```shell
./store --port <port> --depth <depth>
```

`<depth>` will be the number of keys the store can retain. If the store is
full when a `PUT` is made then the key that was least recently used will be
evicted _regardless of owner_. A key is considered used if it is written,
via `PUT`, or read via `GET`.

### Enhanced List

> Capability: `list`

Update `/list` and `/list/<key>` to return the following for each key:

```
{
  "key": "<key>",
  "owner": "<user>",
  "writes": <number of times written>,
  "reads": <numer of reads>,
  "age": <milliseconds since last read or write>
}
```

## Testing

See [harness/README.me]().

## Review

Your code will be reviewed to ensure requirements not covered by the test
harness are met. A template `.golangci.yml` file for [golangci-lint][] is
provided to help ensure good coding practices.

[golangci-lint]: https://github.com/golangci/golangci-lint

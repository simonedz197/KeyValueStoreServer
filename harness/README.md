# Test Harness

To run the harness use:

```shell
go run ./...
```

You will be prompted for the URL to access your key/value store and the set of
extra capabilities you may have added and want testing. These are

`login` - you have implemented the `/login` endpoint and the harness should use 
the credentials it provides, as well as test logging in with invalid 
credentials.

`override` - the `admin` role can PUT and DELETE items others have already 
added.

`lru` - you have implemented a "Least Recently Used" algorithm. The harness will
prompt for the depth of the LRU cache and use this to test it.

`list` - you have implemented the enhanced `/list` endpoint. The harness will 
check the values returned.

The output will look similar to:

```
Go Academy (Week 2) Test Harness
================================

Please make sure your key/value store is running and
accessible from this machine.

If you have attempted one or more of the stretch goals
the test harness can test them for you. List the 
capabilities you have added as a space or comma separated
list. Valid options are: login, override, lru and list.

URL for your store (default: http://localhost:8000) 
Store capabilities (default: none) lru
LRU depth: 50
Testing /ping                                      ... PASSED
Testing /login                                     ... SKIPPED
Testing PUT/GET/DELETE /store/<key>                ... PASSED
Checking user can alter another users data         ... PASSED
Checking output from /list                         ... SKIPPED
Testing admin override                             ... SKIPPED
Testing LRU cache                                  ... PASSED
Testing /shutdown                                  ... PASSED

All tests passed, well done!
```

## Tweaking The Tests

[variables.go]() contains a few levers that you can play with, mainly the
default values when running the tests.

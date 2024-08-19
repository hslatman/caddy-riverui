# caddy-riverui

A POC [Caddy](https://caddyserver.com/) module serving [RiverUI](https://github.com/riverqueue/riverui)

## Description

This module is currently WIP. Things will change.

It relies on the [RiverUI](https://github.com/riverqueue/riverui) library, which itself is also under development.



## Usage

Create a (custom) Caddy server (or use xcaddy)

```golang
package main

import (
  cmd "github.com/caddyserver/caddy/v2/cmd"
  _ "github.com/caddyserver/caddy/v2/modules/standard"

  // enable the RiverUI handler
  _ "github.com/hslatman/caddy-riverui"
)

func main() {
  cmd.Main()
}
```

Example Caddyfile (without route):

```text
{
    order riverui first
}

localhost {
    riverui
}
```

Example Caddyfile (with route):

```text
localhost {
    route {
        riverui
    }
}
```

## TODO

* Support configuration through Caddyfile, JSON and environment
* Make configuration compatible with RiverUI (e.g. RIVER_DEBUG)
* Fix CORS configuration option and ensure working as expected
* Ensure handler is provisioned lazily (i.e. no active DB connection establishment)
* Support running on host/port other than https://localhost:80, which is currently hardcoded in the web app build
* Split frontend and backend handlers?
* Example with authentication, using caddy-security (AuthCrunch)?
* ...
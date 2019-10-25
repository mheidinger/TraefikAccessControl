# TraefikAccessControl

<img align="right" width="180px" src="logo.png">

#### TraefikAccessControl is a small server application providing a ForwardAuth source for Traefik and is configurable on a per user and URL basis.

|                          |                                |
:-------------------------:|:-------------------------------:
![](screenshots/login.png) |  ![](screenshots/dashboard.png)

## Installation

### Docker

Docker images are available on [Docker Hub](https://hub.docker.com/r/mheidinger/traefik-access-control).

The default database path inside the docker container is `/app/tac.db`.

### Compiling and running locally

For compiling the source, first download all dependencies:
```
go mod download
```

After that the compilation and execution can be done via the Makefile:
```
make
make run
```

To generate and automatically import some test data (generated in `testData.go`) run the following:
```
make run-import
```

A cleanup of the database and the generated test data is also possible:
```
make clean
```

## Usage

```
./TraefikAccessControl -help

Usage of ./TraefikAccessControl:
  -cookie_name string
        Cookie name used (default "tac_token")
  -db_name string
        Path of the database file (default "tac.db")
  -force_import
        Force the import of the given file, deletes all existing data
  -import_name string
        Path of an file to import
  -port int
        Port on which the application will run (default 4181)
  -user_header_name string
        Header name that contains the username after successful auth (default "X-TAC-User")
```

If at the start of TAC no user exists, a new admin user will be created.
The credentials for this user are printed to the logs.

### Traefik configuration (version 2.0)

```
[http.middlewares.tac-auth.forwardAuth]
	address = "https://your_tac_url/access"
	authResponseHeaders = ["X-TAC-User"]
```

This configuration will forward the header (`-user_header_name`) with the username to the requested service. 

If you are deploying TAC as its own service, something like the following configuration is needed to ensure that Traefik forwards all needed headers to TAC.
Otherwise the forwarded request will go through Traefik again to reach TAC but all `X-Forwarded-` Headers will be stripped and TAC can't function.
```
[entrypoints.https]
	address = ":443"
	[entrypoints.https.forwardedHeaders]
		trustedIPs = ["127.0.0.1/32", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fd00::/8"]
```
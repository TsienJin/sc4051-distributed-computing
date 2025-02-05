# SERVER

This directory contains the server code for SC4051's Y24/25-S2 project.
The server is built with Go `1.23` and is deployed via Docker to a remote server.

---

## Getting Started

### Tooling

1. [Task](https://taskfile.dev/)
2. [Go](https://go.dev/)
3. [Docker](https://www.docker.com/)

### Running the Server

The server can be run locally or deployed to a remote server; the `Taskfile.yml` assists with the necessary commands to start the server.

1. Running the server locally
```shell
task start
```

2. Running the server locally with Docker
```shell
task start:docker
```

3. Deploy server to remote server using Docker
```shell
task prod
```

4. Watch terminal output from server using `netcat` over a TCP socket connection.
```shell
task prod:w
```

5. Watch terminal output from server using Docker. This requires SSH access to the host machine.
```shell
task prod:w:docker
```

> [!IMPORTANT]
> Task commands that involve Docker (i.e. commands 2, 4, 5) require the environment file `Dockercompose.env` to
> be present. Do reference `Dockercompose.env.sample` for the necessary environment variables to be defined.

> [!IMPORTANT]
> Task commands that involve the remote server (i.e. commands 3, 4, 5) require the environment file `Taskfile.env` to
> be present. Do reference `Taskfile.env.sample` for the necessary environment variables to be defined.

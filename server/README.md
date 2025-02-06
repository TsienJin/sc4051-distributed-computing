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
Alternatively, the following command can be executed using `netcat` to achieve the same outcome. This is useful when the
environment variables are not set.
```shell
nc -v <host> <port>
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

---

## Environment Variables

Environment variables are used to configure a handful of system level behaviours, ranging from port allocation
to packet drop rate and intervals between clean-ups.

### `Dockercompose.env`

1. `SERVER_PORT` -- Port exposed to UDP for connections.
2. `SERVER_LOG_PORT` -- Port exposed for watching server logs (and sending client logs).
3. `PACKET_DROP_RATE` -- [0,1] Rate of which incoming and outgoing packets are dropped.
4. `PACKET_TIMEOUT_RECEIVE` -- Minimum time before packet is requested again.
5. `MESSAGE_ASSEMBLER_INTERVAL` -- Time interval (in milliseconds) that partial messages are checked for missing packets.
6. `RESPONSE_TTL` -- Time (in milliseconds) that sent responses are kept on the server.
7. `RESPONSE_INTERVAL` -- Time (in milliseconds) that the system checks for "expired" responses.

### `Taskfile.env`

1. `REMOTE_HOST` -- Server address, used for SSH etc.
2. `REMOTE_USER` -- Server user, used for SSH when building to remote server using Docker compose.

---


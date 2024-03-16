# honeytokenWebserver

## Introduction

Simple honeytoken server that handles HTTP URL based honeytokens.

## Acknowledgements

Inspired by <https://canarytokens.org>, but was looking for something simpler and easier to maintain in a selfhosted environment.

## Basic Setup

### Build

Prerequisites:
- You need a working golang installation to compile the server. 

Executing `make` in the source directory will build a Linux and a Windows executable:

```
> make
go fmt ./...
go vet ./...
GOOS=linux GOARCH=amd64 go build -tags netgo .
GOOS=windows GOARCH=amd64 go build .
```

Result:

```
-rwxr-xr-x 1 user users 16499129 16. Mär 14:54 honeytokenWebServer
-rwxr-xr-x 1 user users 16807936 16. Mär 14:54 honeytokenWebServer.exe
```

### Configuration

honeytokenWebServer needs a config file in JSON format and a SQLite database to work.

#### Config File

You can find a example of the JSON config file in `config.json_example`.

Required settings are:

- `interfaceAndPort`: Interface and port the server will bind to, e.g. `"localhost:20000"` or `":20000"`
- `responseFile`: File in local filesystem of the server, that contains the HTTP response body that will be sent to clients for every request, e.g. `"response.txt"`
- `responseContentType`: HTTP content-type header that will be sent to clients for every request, e.g. `"text; charset=UTF-8"`
- `responseCode`: HTTP response code that will be sent to clients for every request, e.g. `200`
- `sqliteDatabase`: SQlite database containing the configured honeytokens, e.g. `"honeyDB.sqlite"`
- `smtpServer`: DNS name of SMTP server, e.g. `"smtp.example.com"`
- `smtpPort`: SMTP server port, e.g. `587`
- `smtpUser`: Username and senders email, e.g. `"honeytokenserver@example.com"`
- `smtpPassword`: Password of the `smtpUser`

Optional settings are:

- verbose: `true` for verbose logging, `false` for standard logging

#### SQLite database

SQLite database containing the configured honeytokens

The configured SQLite database file will be created automatically if not existing.

Every entry represents an honeytoken based on a specific URL of the server, e.g. if `url` columen contains `"/abc"` and your server is reachable via `https://web.example.com` the honeytoken URL the client would have to trigger is `https://web.example.com/abc`.

The following columnes need to be filled:

- `url`: See above.
- `title`: A title you can freely select
- `comment`: A description you can freely select. Recommendation is to describe in detail where this Honeytoken has been deployed, so you are able to tell which system/file/dataset has been compromised.
- `notify_receiver`: E-Mail address of the intended receiver of the alert, e.g. `alerts@example.com`.

You can use any SQLite client to edit the database, e.g. <https://sqlitebrowser.org/>.

### Run Server

Execute the server on Linux via

    .\honeytokenWebServer -configFile config.json

on Windows via

    ./honeytokenWebServer.exe -configFile config.json

Output would be e.g.

```
Starting honeytokenWebServer with interface "localhost:20000" response file "response.txt" and response Content-Type "text; charset=UTF-8"
```

Note: It is recommended to use a reverse proxy (e.g. [Traefik](https://traefik.io/traefik/) or [NGINX](https://www.nginx.com/)) to provide services via HTTPS and allow multiple services on the same server via the same port.

## Deployment via Ansible / Docker

You can easily deploy the Honeytoken web server using Docker and Ansible.

Some examples to get you started, using a Traefik proxy with TLS using Lets Encrypt (the Traefik installation and configuration is not included):

A sample Dockerfile for creating the Honeytoken web server container:

```
FROM gcr.io/distroless/static
LABEL maintainer="honeytokens"
USER nonroot:nonroot
COPY --chown=nonroot:nonroot honeytokenWebServer /
LABEL traefik.enable=true
LABEL traefik.http.routers.honey.rule=Host(`web.example.com`)
LABEL traefik.http.routers.honey.entrypoints=websecure
LABEL traefik.http.routers.honey.tls=true
LABEL traefik.http.routers.honey.tls.certresolver=letsencrypt
LABEL traefik.http.services.honey.loadBalancer.server.port=20000
CMD ["/honeytokenWebServer","-configFile","/config/config.json"]
```

In this case the JSON config file must be mapped as volume from the servers filesystem to the container at `/config`.

The following Ansible tasks file installs the docker container on a Linux server:

```
---

# this tasks expect a working docker and traefik installation (traefik with `web` network)

- name: Create honeytokenWebServer folder
  become: true
  ansible.builtin.file:
    path: /opt/honeytokenWebServercontainer
    state: directory
    mode: '0755'
    
- name: Create honeytokenWebServer/config folder
  become: true
  ansible.builtin.file:
    path: /opt/honeytokenWebServercontainer/config
    state: directory
    mode: '0755'
    
- name: Copy honeytokenWebServer binary
  become: true
  copy:
    src: honeytokenWebServer
    dest: /opt/honeytokenWebServercontainer/honeytokenWebServer
    owner: root
    group: root
    mode: '0755'

- name: Copy honeytokenWebServer Token DB
  become: true
  copy:
    src: honeyDB.sqlite
    dest: /opt/honeytokenWebServercontainer/config/honeyDB.sqlite
    owner: root
    group: root
    mode: '0655'

- name: Copy honeytokenWebServer config.json
  become: true
  copy:
    src: config.json
    dest: /opt/honeytokenWebServercontainer/config/config.json
    owner: root
    group: root
    mode: '0655'

- name: Copy honeytokenWebServer response.txt
  become: true
  copy:
    src: response.txt
    dest: /opt/honeytokenWebServercontainer/config/response.txt
    owner: root
    group: root
    mode: '0655'

- name: Copy Dockerfile
  become: true
  copy:
    src: Dockerfile
    dest: /opt/honeytokenWebServercontainer/Dockerfile
    owner: root
    group: root
    mode: '0655'

# remove old container, if exist
- name: Tear down existing honeytokenWebServer containers
  community.docker.docker_container:
    name: honeytokenwebserver
    state: absent
    container_default_behavior: no_defaults

# remove old image, if exists, to enforce updates
- name: Remove honeytokenwebserver image
  docker_image:
    state: absent
    name: honeytokenwebserver:v1.0

# remove old image, if exists, to enforce updates
- name: Remove distroless image
  docker_image:
    state: absent
    name: gcr.io/distroless/static

# docker build -t honeytokenwebserver .
- name: Build honeytokenwebserver container
  community.docker.docker_image:
    name: honeytokenwebserver:v1.0
    build: 
      path: /opt/honeytokenWebServercontainer
    source: build
    state: present

# docker run --restart always --name honeytokenwebserver --network=web -d honeytokenwebserver [...]
- name: Run the honeytokenWebServer container
  community.docker.docker_container:
    name: honeytokenwebserver
    image: honeytokenwebserver:v1.0
    state: started
    restart: yes
    restart_policy: unless-stopped
    container_default_behavior: no_defaults
    volumes:
      - "/opt/honeytokenWebServercontainer/config:/config"
    network_mode: default
    networks:
      - name: "web"
```

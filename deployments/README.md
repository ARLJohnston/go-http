# Deployments

## Docker Compose
To run the whole application:
```console
docker compose up -d
```
Note that this will build the front-end and back-end from local changes.

To shutdown and wipe database:
```console
docker compose down -v
```

## Docker Swarm
To deploy to a stack named <NAME>:
```console
docker stack deploy -c docker-swarm.yml <NAME>
```
This will pull the images from their remotes.

To remove the stack:
```console
docker stack rm <NAME>
```

Note that these may require root, if Docker has not been setup for rootless access.

# Kubernetes
To deploy the application to a running cluster
```console
cd kompose
kubectl apply -f .
```

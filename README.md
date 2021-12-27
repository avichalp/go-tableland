# Go Tableland

This project is part of a POC of the Tableland project.

![Tableland POC](https://user-images.githubusercontent.com/1233473/147493247-4710159a-86f3-4e80-8e4e-36ba7499eafc.png)

It implements the validator as a JSON-RPC server responsible for updating a Postgres database.

## API Spec

[Postman Collection](https://www.postman.com/aviation-participant-86342471/workspace/my-workspace/collection/18493329-068ef574-afde-4057-926c-ebee6628315c)

## Current state of the project (and design decisions)

- It has a JSON-RPC server responsible for handling user calls such as `createTable` and `runSQL`
- It has additional HTTP endpoints for getting data from the sytems
  - `GET /tables/{uuid}` for table's metadata
  - `GET /tables/controller/{address}` to get all tables a controller address owns
- It uses JWT authentication on the JSON-RPC calls. The JWT is token is created in the client using [textileio/storage-js](https://github.com/textileio/storage-js/blob/main/packages/eth/src/index.ts#L66)
- The JSON-RPC is implemented using Ethereum's [implementation](https://pkg.go.dev/github.com/ethereum/go-ethereum/rpc) of the [2.0 spec](https://www.jsonrpc.org/specification)
- The JSON-RPC server is an HTTP server (Just for clarification. It could also be just a TCP server.)
- The server is currently deployed as a `docker` container inside a [Compute Engine VM](https://console.cloud.google.com/compute/instances?project=textile-310716&authuser=1)
- Configs can be passed with flags, config.json file or env variables (it uses the [uConfig](https://github.com/omeid/uconfig) package)
- There is a Postgres database running inside the same [Compute Engine VM](https://console.cloud.google.com/compute/instances?project=textile-310716&authuser=1) as the container
- For local development, there is a `docker-compose` file. Just execute `make up` to have the validator up and running.
- For local development, you need to connect to a Ethereum node. The best approach now is to run a local `hardhat` container using [textileio/eth-tableland/pull/5](https://github.com/textileio/eth-tableland/pull/5)

## How to publish a new version

This project uses `docker` and Google's [Artifact Registry](https://console.cloud.google.com/artifacts?authuser=1&project=textile-310716) for managing container images.

```bash
make image    # builds the image
make publish  # publishes to Artifact Registry
```

Make sure you have `gcloud` installed and configured.
If you get an error while trying to publish, try to run `gcloud auth configure-docker us-west1-docker.pkg.dev`

## How to deploy

```bash
docker run -d --name api -p 80:8080 --add-host=database:172.17.0.1 -e DB_HOST=database -e DB_PASS=[[PASSWORD]] -e DB_USER=validator -e DB_NAME=tableland -e DB_PORT=5432 -e REGISTRY_ETHENDPOINT=http://tableland.com:8545 -e REGISTRY_CONTRACTADDRESS=0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512 [[IMAGE]]
```

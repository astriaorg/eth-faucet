default:
  @just --list

set dotenv-filename := ".env.local"

# installs developer dependencies for MacOS users with `brew`
brew-install-dev-deps:
  brew install just
  brew install mprocs
  brew install golangci-lint
  brew install watchexec

# installs deps for front end app
web-install-deps:
  cd web && npm install

# runs the full web app. generates the front end app before starting the server.
run:
  go generate -x
  go run -v ./... -httpport 8089 -firestoreprojectid $FIRESTORE_PROJECT_ID

# run cli and restart when code changes
run-watch:
  watchexec --exts go,mod,svelte,js,ts,html --clear --restart -- just run

# lints the go code
go-lint:
  golangci-lint run ./...

# tests the go code
go-test:
  go test -v ./...

# run go tests whenever code changes
go-test-watch:
  @just _watch-go go-test

# formats Go code
go-fmt:
  go fmt ./...

# prettifies web code
web-fmt:
  cd web && npm run prettier

## Docker

image_name := 'ghcr.io/astriaorg/eth-faucet'
default_docker_tag := 'local'

# builds docker image w/ `local` tag by default
docker-build tag=default_docker_tag:
  docker buildx build -f ./Dockerfile -t {{image_name}}:{{tag}} .

# runs faucet via docker
docker-run tag=default_docker_tag:
  docker run --rm -p 8080:8080 {{image_name}}:{{tag}}

## Helpers

# run `command` whenever Go code changes
_watch-go command:
  watchexec --exts go,mod --clear --restart -- just {{command}}

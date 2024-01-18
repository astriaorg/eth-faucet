default:
  @just --list

image_name := 'ghcr.io/astriaorg/eth-faucet'
default_docker_tag := 'local'

# installs developer dependencies for MacOS users with `brew`
brew-install-dev-deps:
  brew install just
  brew install mprocs
  brew install golangci-lint
  brew install watchexec

# builds front end web app and go binary
build-all:
  go generate -x
  go build -v

# builds docker image w/ `local` tag by default
docker-build tag=default_docker_tag:
  docker buildx build -f ./Dockerfile -t {{image_name}}:{{tag}} .

# runs faucet via docker
docker-run tag=default_docker_tag:
  docker run --rm -p 8080:8080 {{image_name}}:{{tag}}

# lints the go code
go-lint:
  golangci-lint run ./...

# tests the go code
go-test:
  go test -v ./...

# run go tests whenever code changes
go-test-watch:
  @just _watch-go go-test

# runs the go binary
go-run:
  go run -v ./... -httpport 8080

# run cli and restart when code changes
go-run-watch:
  @just _watch-go go-run

# formats Go code
go-fmt:
  go fmt ./...

# installs deps for front end app
web-install-deps:
  cd web && npm install

# prettifies web code
web-fmt:
  cd web && npm run prettier

# run the front end dev server
web-run:
  cd web && npm run dev

# run front end and backend via mprocs (`brew install mprocs` may be needed)
run-all-dev:
  mprocs "just web-run" "just go-run"


## Helpers

# run `command` whenever Go code changes
_watch-go command:
  watchexec --exts go,mod --clear --restart -- just {{command}}

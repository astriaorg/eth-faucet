default:
  @just --list

# builds front end web app and go binary
build:
  go generate -x
  go build -v

# lints the go code
go-lint:
  golangci-lint run ./...

# tests the go code
go-test:
  go test -v ./...

# runs the go binary
go-run:
  go run -v ./... -httpport 8080

# installs deps for front end app
web-install-deps:
  cd web && npm install

# prettifies web code
web-prettier:
  cd web && npm run prettier

# run the front end dev server
web-run:
  cd web && npm run dev

# run front end and backend via mprocs (`brew install mprocs` may be needed)
run-all-dev:
  mprocs "just web-run" "just go-run"

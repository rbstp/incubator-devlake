### Build all .so files for plugins
```
rm -rf backend/bin/plugins

docker run --rm --platform=linux/amd64 \
  -v "$PWD/backend":/workspace \
  -w /workspace \
  golang:1.22-bookworm \
  bash -lc '
    set -euo pipefail
    export PATH=/usr/local/go/bin:$PATH
    go version
    apt-get update && apt-get install -y \
      gcc binutils libssh2-1-dev libssl-dev zlib1g-dev pkg-config
    go mod download
    mkdir -p bin/plugins/github bin/plugins/gitlab bin/plugins/jira bin/plugins/refdiff \
             bin/plugins/issue_trace bin/plugins/dora bin/plugins/org bin/plugins/slack
    export CGO_ENABLED=1 GOOS=linux GOARCH=amd64
    go build -buildmode=plugin -o bin/plugins/github/github.so ./plugins/github
    go build -buildmode=plugin -o bin/plugins/gitlab/gitlab.so ./plugins/gitlab
    go build -buildmode=plugin -o bin/plugins/jira/jira.so ./plugins/jira
    go build -buildmode=plugin -o bin/plugins/refdiff/refdiff.so ./plugins/refdiff
    go build -buildmode=plugin -o bin/plugins/issue_trace/issue_trace.so ./plugins/issue_trace
    go build -buildmode=plugin -o bin/plugins/dora/dora.so ./plugins/dora
    go build -buildmode=plugin -o bin/plugins/org/org.so ./plugins/org
    go build -buildmode=plugin -o bin/plugins/slack/slack.so ./plugins/slack
  '
```

### Build just the tested plugins
```
docker run --rm --platform=linux/amd64 \
  -v "$PWD/backend":/workspace \
  -w /workspace \
  golang:1.22-bookworm \
  bash -lc '
    set -euo pipefail
    export PATH=/usr/local/go/bin:$PATH
    go version
    apt-get update && apt-get install -y \
      gcc binutils libssh2-1-dev libssl-dev zlib1g-dev pkg-config
    go mod download
    mkdir -p bin/plugins/slack
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
      go build -buildmode=plugin -o bin/plugins/slack/slack.so ./plugins/slack
  '
```

### Build Docker images and run with docker-compose
```
docker build --platform=linux/amd64 -f Dockerfile.devlake-official-style -t devlake-from-source-official:latest .
docker build --platform=linux/amd64 -f config-ui/Dockerfile -t devlake-config-ui-test:latest config-ui/


docker-compose -f docker-compose-test-new.yml up -d
```
### Tear down
```
docker-compose -f docker-compose-test-new.yml down
```

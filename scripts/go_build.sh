#! /usr/bin/env bash
set -eu -o pipefail
_wd=$(pwd)
_path=$(dirname $0 | xargs -i readlink -f {})

Program="crons"

buildTime=$(date +'%FT%T%:z')
gitBranch="$(git rev-parse --abbrev-ref HEAD)" # current branch
gitCommit=$(git rev-parse --verify HEAD) # git log --pretty=format:'%h' -n 1
gitTime=$(git log -1 --format="%at" | xargs -I{} date -d @{} +%FT%T%:z)
gitTreeState="clean"

uncommitted=$(git status --short)
unpushed=$(git diff origin/$gitBranch..HEAD --name-status)
[[ ! -z "$uncommitted$unpushed" ]] && gitTreeState="dirty"

ldflags="\
  -X main.buildTime=${buildTime} \
  -X main.gitBranch=$gitBranch   \
  -X main.gitCommit=$gitCommit   \
  -X main.gitTime=$gitTime       \
  -X main.gitTreeState=$gitTreeState"

mkdir -p target
go build -ldflags="$ldflags" -o target/$Program main.go
GOOS=linux   GOARCH=amd64 go build -o target/${Program}     -ldflags="-w -s $ldflags" main.go
GOOS=windows GOARCH=amd64 go build -o target/${Program}.exe -ldflags="-w -s $ldflags" main.go
ls -lh target/${Program}*

#!/usr/bin/env bash
# Local dev: compiles every component first (stops on the first failure), then runs them all
# in this terminal with prefixed logs. Ctrl-C stops everything.
# Usage: ./dev.sh            # all components
#        ./dev.sh engine ui  # only the ones you name (ui | api | engine)
set -euo pipefail
cd "$(dirname "$0")"
ROOT=$PWD

components=("$@")
[ ${#components[@]} -eq 0 ] && components=(engine api ui)
for c in "${components[@]}"; do
  case $c in
    ui|api|engine) ;;
    *) echo "unknown component: $c (use ui, api, engine)"; exit 1 ;;
  esac
done

build_engine() {
  protoc -I proto \
    --go_out=engine-go/finance --go_opt=paths=source_relative \
    --go-grpc_out=engine-go/finance --go-grpc_opt=paths=source_relative \
    proto/finance.proto
  (cd engine-go && go build -o bin/server ./cmd/server)
}

build_api() {
  # Maven regenerates Java gRPC + OpenAPI sources from proto during the build.
  mvn -q -pl api-spring -am clean package -DskipTests
}

build_ui() {
  (cd finankal-ui && { [ -d node_modules ] || npm install; } && npm run build)
}

for c in "${components[@]}"; do
  echo "== [$c] compiling"
  "build_$c"
done
echo "== build ok, starting components"

# run <name> <dir> <command> in the background, each log line prefixed with [name].
run() {
  ( cd "$ROOT/$2" && bash -c "$3" 2>&1 | sed -u "s/^/[$1] /" ) &
}

set +e
# Ctrl-C (or script exit) stops every component.
trap 'trap - INT TERM EXIT; kill 0' INT TERM EXIT
for c in "${components[@]}"; do
  case $c in
    engine) run engine engine-go "./bin/server" ;;
    api)    run api . "java -jar $(ls api-spring/target/*.jar | grep -v plain | head -1) --spring.profiles.active=local" ;; # local profile sets grpc.client.host
    ui)     run ui finankal-ui "npm run dev" ;; # vite dev keeps hot reload
  esac
done
wait

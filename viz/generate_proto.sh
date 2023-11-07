protoc -I=. \
    --plugin=protoc-gen-ts=./viz/node_modules/.bin/protoc-gen-ts \
    --ts_out=. \
    ./viz/src/dashboard_state.proto
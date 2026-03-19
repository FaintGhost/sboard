package rpc

//go:generate sh -c "cd ../.. && buf generate && buf generate --template buf.node.gen.yaml --path proto/sboard/node/v1/node.proto && cd ../web && bunx oxfmt --write src/lib/rpc/gen"

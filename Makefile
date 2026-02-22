.PHONY: generate generate-go generate-ts check-generate

generate: generate-go generate-ts

generate-go:
	cd panel && go generate ./internal/api/...

generate-ts:
	cd panel/web && bun run generate

check-generate: generate
	@if ! git diff --exit-code -- '*.gen.go' '*.gen.ts'; then \
		echo ""; \
		echo "ERROR: Generated files are out of date."; \
		echo "Run 'make generate' and commit the changes."; \
		exit 1; \
	fi
	@echo "Generated files are up to date."

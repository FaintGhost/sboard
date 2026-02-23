.PHONY: generate generate-go generate-ts check-generate e2e e2e-smoke e2e-down e2e-report

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

e2e:
	cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright

e2e-smoke:
	cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
	docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke ; \
	ret=$$? ; \
	docker compose -f docker-compose.e2e.yml down -v ; \
	exit $$ret

e2e-down:
	cd e2e && docker compose -f docker-compose.e2e.yml down -v

e2e-report:
	cd e2e && bunx playwright show-report playwright-report

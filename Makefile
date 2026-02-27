.PHONY: generate generate-rpc check-generate e2e e2e-smoke e2e-down e2e-report

generate: generate-rpc

generate-rpc:
	cd panel && go generate ./internal/rpc/...

check-generate: generate
	@if ! git diff --exit-code -- 'panel/internal/rpc/gen/**' 'panel/web/src/lib/rpc/gen/**'; then \
		echo ""; \
		echo "ERROR: Generated files are out of date."; \
		echo "Run 'make generate' and commit the changes."; \
		exit 1; \
	fi
	@echo "Generated files are up to date."

e2e:
	cd e2e && docker compose -f docker-compose.e2e.yml down -v && \
	docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright ; \
	ret=$$? ; \
	docker compose -f docker-compose.e2e.yml down -v ; \
	exit $$ret

e2e-smoke:
	cd e2e && docker compose -f docker-compose.e2e.yml down -v && \
	docker compose -f docker-compose.e2e.yml up --build -d panel node && \
	docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke ; \
	ret=$$? ; \
	docker compose -f docker-compose.e2e.yml down -v ; \
	exit $$ret

e2e-down:
	cd e2e && docker compose -f docker-compose.e2e.yml down -v

e2e-report:
	cd e2e && bunx playwright show-report playwright-report

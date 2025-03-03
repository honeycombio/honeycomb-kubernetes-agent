MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-builtin-variables

.PHONY: install-tools
install-tools:
	go get -tool github.com/google/go-licenses/v2@v2.0.0-alpha.1
	go mod tidy

.PHONY: update-licenses
update-licenses: install-tools
	rm -rf LICENSES; \
	go tool go-licenses save . --save_path LICENSES;

.PHONY: verify-licenses
verify-licenses: install-tools
	go tool go-licenses save . --save_path temp; \
    if diff temp LICENSES > /dev/null; then \
      echo "Passed"; \
      rm -rf temp; \
    else \
      echo "LICENSES directory must be updated. Run make update-licenses"; \
      rm -rf temp; \
      exit 1; \
    fi; \

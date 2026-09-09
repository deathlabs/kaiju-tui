# ---------------------------------------------------------
# Set the default target.
# ---------------------------------------------------------

.DEFAULT_GOAL := build

# ---------------------------------------------------------
# Set the default values.
# ---------------------------------------------------------
VERSION := $(shell date +%Y-%m-%d-%H%M%S)
SBOM_FILE_JSON ?= kaiju-tui-sbom.json
VEX_FILE_YAML ?= vex.yaml
VEX_FILE_JSON ?= vex.json
VEX_AUTHOR ?= Victor Fernandez III
VEX_ID_BASE ?= kaiju-tui

# Security scanner configurations.
SEMGREP_CONFIG ?= auto
GRYPE_FAILURE_THRESHOLD ?= medium

# ---------------------------------------------------------
# Check for bugs.
# ---------------------------------------------------------

.PHONY: check
.SILENT: check

check:
	golangci-lint run .

# ---------------------------------------------------------
# Check the source code for vulnerabilities.
# ---------------------------------------------------------

.PHONY: sast
.SILENT: sast

sast:
	semgrep scan --config $(SEMGREP_CONFIG) .

# ---------------------------------------------------------
# Generate VEX statements.
# ---------------------------------------------------------

.PHONY: vex
.SILENT: vex

define VEX_FILTER
{
  "@context": "https://openvex.dev/ns/v0.2.0",
  "@id": "$(VEX_ID_BASE)-" + now,
  "author": "$(VEX_AUTHOR)",
  "timestamp": now,
  "version": 1,
  "statements": [
    .advisories[] | {
      "vulnerability": { "name": .vulnerability },
      "products": [ .products[] | { "@id": . } ],
      "status": .status,
      "justification": .justification,
      "impact_statement": .impact_statement
    }
  ]
}
endef
export VEX_FILTER

vex:
	if [ -f "$(VEX_FILE_YAML)" ]; then \
		yq -o=json "$$VEX_FILTER" "$(VEX_FILE_YAML)" > "$(VEX_FILE_JSON)"; \
	else \
		rm -f "$(VEX_FILE_JSON)"; \
	fi

# ---------------------------------------------------------
# Generate SBOMs for the container images.
# ---------------------------------------------------------

.PHONY: sbom
.SILENT: sbom

sbom: 
	syft . -o cyclonedx-json=$(SBOM_FILE_JSON)

# ---------------------------------------------------------
# Scan dependencies for vulnerabilities.
# ---------------------------------------------------------

.PHONY: dependency-scan
.SILENT: dependency-scan

dependency-scan: sbom vex
	grype db update &&\
	if [ -f "$(VEX_FILE_JSON)" ]; then \
		grype sbom:$(SBOM_FILE_JSON) --vex $(VEX_FILE_JSON) --fail-on $(GRYPE_FAILURE_THRESHOLD); \
	else \
		grype sbom:$(SBOM_FILE_JSON) --fail-on $(GRYPE_FAILURE_THRESHOLD); \
	fi

# ---------------------------------------------------------
# Build the artifact.
# ---------------------------------------------------------

.PHONY: build
.SILENT: build

build: dependency-scan
	go install -ldflags="-s -w -X 'github.com/deathlabs/kaiju-tui/cmd.version=$(VERSION)'" . &&\
	echo "Built and installed kaiju-tui v$(VERSION)."

# ---------------------------------------------------------
# Update dependencies.
# ---------------------------------------------------------

.PHONY: update
.SILENT: update

update:
	go get -u ./...
	go mod tidy

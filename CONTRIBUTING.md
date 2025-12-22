Contributing to Helix CLI
=========================

So you want to improve the tool that builds our systems. Good. But before you push code, understand the rules.

The Philosophy
--------------

1.  **Boring is Better:** We optimize for maintenance, not for "clever" code. If your PR makes the generated code harder to understand for a junior engineer during an outage, it will be rejected.

2.  **Explicit > Implicit:** No magic auto-wiring. No hidden reflection hell. The generated code must be readable and explicitly defined.

3.  **Survivability:** Prioritize failure handling, retries, and observability features over syntactic sugar.

Getting Started
---------------

1.  Clone the repo.

2.  Install dependencies: `go mod download`.

3.  Run the CLI locally: `go run main.go`.

Development Workflow
--------------------

### Testing Templates

Since this is a generator, testing is tricky.

1.  Make changes to files in `templates/`.

2.  Build the CLI: `go build -o helix-cli .`

3.  Run it against a temporary directory: `./helix-cli init svc-test-gen`.

4.  **Crucial:** Navigate into `svc-test-gen` and run `make init`. If the generated code doesn't compile or the tests fail, your PR is broken.

### Adding New Features

-   **Discuss first:** Open an issue. Explain the *business value* of the new feature.

-   **Keep it modular:** If you add a new "adapter" (e.g., RabbitMQ), ensure it doesn't pollute the core logic.

Pull Request Checklist
----------------------

-   \[ \] Does the generated code compile?

-   \[ \] Does `golangci-lint` pass on the generated code?

-   \[ \] Have you updated `README.md` if CLI args changed?

-   \[ \] Did you bump the version in `main.go` (if applicable)?

Reporting Bugs
--------------

Don't just say "it broke".

1.  Paste the command you ran.

2.  Paste the error log.

3.  Paste the OS/Go version.

**Ship it or fix it.**
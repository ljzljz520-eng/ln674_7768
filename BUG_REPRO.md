# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	campus-device-hub/cmd/hub	[no test files]
?   	campus-device-hub/internal/retry	[no test files]
--- FAIL: TestWorkflowConcurrentConfirmation (0.00s)
    workflows_test.go:87: confirmation workflow lost state: {ID:sync:workflow Vendor:BrightGrid StartedAt:0001-01-01 00:00:00 +0000 UTC FinishedAt:0001-01-01 00:00:00 +0000 UTC State:success DeviceCount:0 ErrorMessage: Confirmations:1 AuditEventIDs:[confirm:sync:workflow:operator-b] Revision:0}
FAIL
FAIL	campus-device-hub/integration	0.079s
ok  	campus-device-hub/internal/adapters	0.004s
ok  	campus-device-hub/internal/dashboard	0.015s
ok  	campus-device-hub/internal/domain	0.005s
--- FAIL: TestConcurrentConfirmationKeepsBothOperators (0.01s)
    confirm_test.go:51: confirmation count = 1
FAIL
FAIL	campus-device-hub/internal/service	0.056s
ok  	campus-device-hub/internal/storage	0.020s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/hub): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/hub): exit `0`

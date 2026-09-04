# Echo extension example

Build the runner for the platform that will install it, then install the directory:

```bash
go build -o echo-runner .
fnsh ext add .
fnsh example echo hello
```

On Windows, Go emits `echo-runner.exe`; FinishBit resolves the platform suffix automatically, so the manifest remains portable.

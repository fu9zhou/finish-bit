# Extension protocol v1

Extensions are local packages that expose one or more Operations through one executable. The implementation language and bundled runtime are not prescribed.

## Manifest

The extension root must contain `finishbit-extension.json`:

```json
{
  "schema": 1,
  "name": "example-echo",
  "version": "1.0.0",
  "executable": "echo-runner",
  "operations": [
    {
      "id": "example.echo",
      "summary": "Echo an input value",
      "description": "Return the first input unchanged.",
      "aliases": ["echo text"],
      "tags": ["example"],
      "inputs": [{"name": "input", "type": "string", "description": "Value to echo", "required": true}],
      "source": "extension"
    }
  ]
}
```

Install a directory or ZIP whose manifest is at its root:

```bash
fnsh ext add ./example-echo
```

Archives are limited to 512 MiB extracted size. Absolute paths, path traversal, and symbolic links are rejected. Installation does not execute the extension, but invoking one of its Operations does.

## Process exchange

For each invocation FinishBit starts the declared executable, sends one JSON object on stdin, closes stdin, and reads one JSON object from stdout.

Request:

```json
{"protocol":1,"operation":"example.echo","request":{"inputs":["hello"],"options":{}}}
```

Success response:

```json
{"result":{"operation":"example.echo","data":{"text":"hello"}}}
```

Error response:

```json
{"error":{"code":"invalid_input","message":"input is required"}}
```

Diagnostic logs belong on stderr. Stdout is reserved for the single protocol response and limited to 16 MiB. Cancellation terminates the child process through the Go command context.

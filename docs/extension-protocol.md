# Extension protocol v1

Extensions are local packages that expose one or more Operations through one executable. The implementation language and bundled runtime are not prescribed.

The protocol standardizes discovery and message exchange. It does not sandbox the executable, authenticate a publisher, resolve dependencies, or grant capabilities selectively. Install only extensions whose code and provenance you trust.

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

Directory and ZIP installations are limited to 512 MiB total size. Absolute paths, path traversal, and symbolic links are rejected. Installation does not execute the extension, but invoking one of its Operations does.

The manifest `name` identifies the installed extension and `version` describes its release. Operation IDs must follow the same lowercase `<domain>.<action>` rule as core Operations and must not collide with another registered ID. The executable path is relative to the extension root; Windows may resolve the corresponding `.exe` file.

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

The child process inherits the current user's operating-system permissions and environment. Extension authors must not use stdout for logs, progress, or multiple JSON messages.

## Errors

Extensions should return stable, specific error codes and actionable messages. The host treats malformed JSON, oversized output, abnormal process termination, and an invalid response envelope as execution failures. Sensitive values must not be placed in error messages or stderr diagnostics.

## Compatibility

Both the manifest `schema` and request `protocol` are integers. Version 1 implementations must reject unsupported values rather than infer their meaning. Incompatible future shapes will use a new version; see the project [compatibility policy](compatibility.md).

Extension packages should version their own Operation behavior semantically. Removing an Operation, changing a parameter type, or reinterpreting a result can break agent workflows even when the process protocol remains version 1.

## Author checklist

- Keep Operation IDs task-oriented and globally distinctive.
- Validate all inputs and constrain file, memory, subprocess, and output usage.
- Write exactly one response object to stdout and diagnostics only to stderr.
- Exit non-zero when no valid protocol response can be produced.
- Test directory and ZIP installation on every declared platform.
- Document the extension license, runtime requirements, and publisher identity.

See the [example extension](../examples/extensions/echo/README.md), [manifest schema](../schemas/extension-v1.schema.json), and [Operation schema](../schemas/operation-v1.schema.json).

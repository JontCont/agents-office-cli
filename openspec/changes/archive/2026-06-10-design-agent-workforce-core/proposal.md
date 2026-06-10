## Why

Developers currently lack a fast, lightweight, CLI-first way to orchestrate deterministic multi-agent workflows locally with live observability and human intervention. Existing solutions are often web-centric, heavyweight, or optimized for hosted workflows rather than a high-performance local command-line experience with companion monitoring and pause/resume control.

## What Changes

- Implement a deterministic multi-agent Go runtime coordinating a shared discussion thread.
- Establish Go CLI commands to initialize a workspace, execute workflows, launch a companion GUI, and inspect runtime status.
- Implement a decoupled Tauri desktop companion GUI that connects to the CLI over local WebSockets.
- Implement a hybrid supervisor interruption protocol using the state flow: `QUEUED`, `RUNNING`, `INTERRUPTING`, `INTERRUPTED`, `RESUMING`, `COMPLETED`, `FAILED`, `CANCELLED`.
- Support both agent-initiated breakpoints (`ask_human`) and supervisor-initiated on-demand interruption at safe step boundaries.

This MVP is intentionally local-first and single-user. It does not attempt to solve cloud deployment, multi-tenant orchestration, or GUI-based agent configuration.

## Capabilities

### New Capabilities
- agent-workforce: A CLI-first local multi-agent orchestration system with deterministic routing, supervisor interruption, local WebSocket IPC, and a Tauri-based monitoring client.

### Modified Capabilities
- None.

## Impact

### Affected Specs
- openspec/changes/design-agent-workforce-core/specs/agent-workforce/spec.md

### Affected Areas
- CLI entrypoint and command structure
- Workforce runtime and deterministic routing engine
- Event model and interruption state handling
- Local IPC / WebSocket server
- Companion GUI application and monitoring panels

### Expected New Code
- Go CLI and runtime packages
- GUI application scaffolding and monitor components
- Shared event / protocol definitions

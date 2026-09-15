## Journey

<!-- Journey ID and name, e.g. J03 · Categorise a transaction. IDs are stable and never reused -->

## File

<!-- Path to the journey .pen, e.g. apps/web/design/categorise.pen. Never packages/ui/design/product-ui.lib.pen: that file holds reusable mechanics only -->

## Saved to disk

<!-- `just pen-exec` saves on a clean run; `just pen-save` is only for MCP edits to the document a person has open. Confirm git sees the file, then state the mtime or say NOT SAVED -->

## Frames

<!-- One row per frame drawn or changed. Fully qualified name, node ID, zone -->

| Frame | ID | Zone |
| --- | --- | --- |
|  |  |  |

## Surfaces

<!-- For each task surface: inline / modal / drawer / full page, modal or non-modal, route-backed or ephemeral -->

## Components used

<!-- Library components instanced, by name and origin ID in product-ui.lib.pen -->

## Components missing

<!-- Anything the journey needs that the library does not have. Each becomes a task. Name what it is a configuration of, if it is one. Each entry MUST name the existing components you opened and read, and why each cannot carry this - searching by the data it holds, not by the shape you had in mind. Without that line it is a guess, and the usual result is two components owning one job. -->

## States drawn

<!-- Local states beside their step: empty, loading, validation error, failure. Say which were deliberately not drawn and why -->

## Contracts for implementation

<!-- Behaviour the canvas cannot execute: focus management, announcements, routing, keyboard completion. Every number here must have been measured -->

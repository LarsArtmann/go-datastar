# Duplication Baseline (art-dupl)

Committed record of intentionally-accepted clone groups from `art-dupl`
audits. Re-run the canonical invocation below; the actionable output must be
exactly the three groups listed here. A group appearing beyond these three is
a regression to deduplicate; one of these disappearing means the baseline
shrinks — update this file when that happens.

```bash
art-dupl --sort total-tokens -t 1 --suggest-generics --timing --rich-text
# expected: the three groups below (shown count = 3)
```

## Accepted groups

1. `broadcaster := sse.NewBroadcaster[sse.Event]()` —
   `broadcast_example_test.go` (×2) and `example/domain-adapter/main.go`.
   Idiomatic constructor call; a wrapper would hide the instantiated type
   without removing a line.

2. Options-apply loop — `broadcast/options.go` (`newConfig`) vs
   `datastartest/options.go` (`applyRequestOptions`). Cross-module: broadcast
   and datastartest share no importable home, and root's `applyOptions`
   (options.go) is deliberately unexported — it is internal mechanics, not
   protocol API. One fold site per module; a per-module helper used once is
   indirection, not dedup.

3. `if i > 0 { b.WriteString("\n") }` — `datastartest/event.go`
   (`EventsString`) vs `datastartest/require_ordered.go`
   (`renderExpectations`). Canonical separator idiom over differently-shaped
   render loops; a join helper would outweigh the three lines it saves.

## History

- 2026-11-07: actionable groups 7 → 3. Extracted: root options-fold ×5 into
  `applyOptions` (options.go), `CustomEventName` → `quotedSuffix`,
  script open-tag parsing ×2 → `Event.scriptTagParts`
  (datastartest/script_accessors.go), signals guard ×2 →
  `requireSignalsEvent` (datastartest/assert.go).

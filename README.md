# Workbench Line Display — multi-monitor pipeline

Multi-monitor version of the Workbench Display Controller for a **paced assembly
line**. This is a **separate product** from the single-monitor
[`workbench-display`](https://github.com/VisveshRamesh-dev/workbench-display),
which stays frozen for the live production line. Shared heritage, independent
evolution.

## Concept

One scanner feeds a **pipeline of N monitors** on one PC (plus the primary
monitor that runs the plant's own software). Each scan advances the line by one
"tick":

```
         one scanner (entry)
              |  each scan = one tick
              v
   +--------+   +--------+   +--------+
   |Monitor1| ->|Monitor2| ->|Monitor3| -> (complete)
   | Step 1 |   | Step 2 |   | Step 3 |
   +--------+   +--------+   +--------+
   newest       1 scan ago   2 scans ago
```

- Scan product **A** -> Monitor 1 shows **step 1** for A.
- Scan product **B** -> Monitor 1 = step 1 (B), Monitor 2 = step 2 (A).
- Scan product **C** -> M1 = step1 (C), M2 = step2 (B), M3 = step3 (A).

So **Monitor _i_ shows step _i_** for the product that is _i_ positions down the
line, and scanning the next product is what moves the previous one forward.

## Status: IN DESIGN / IN PROGRESS

Forked from the single-monitor `workbench-display` as the shared foundation
(Go HTTP server + `serial-capture` COM reader + SSE display pages). The pipeline
state machine, per-monitor pages, multi-kiosk launcher, and per-step content are
being built on top. **The seeded code still behaves as the single-monitor app**
until the pipeline work lands.

### Open design decisions (finalizing with the client)
1. **Advancement trigger** — each new scan shifts the line forward by one
   (confirmed). (Alternatives considered: a scanner per bench, or a timer.)
2. **Monitor count = step count** — 3–4 monitors mapped to Process 01–04. Define
   behaviour when a part has fewer/more steps than monitors.
3. **Content per monitor** — split each A3 work standard into its per-step panels
   (Process 01/02/03/04) so each monitor shows one step, **vs** show the full
   sheet on every monitor with the current step highlighted.
4. **Mixed parts in the line at once** — front and rear flowing together; each
   monitor shows the step for *its* product.
5. **Edge cases** — duplicate re-scan, "reset line" control, end-of-shift clear,
   supervisor overview of the whole line.

## Planned changes vs the single-monitor base
- **Server:** `stations` -> `displays[]`; a shift-register pipeline (push on
  scan, shift forward, drop off the end); one SSE stream + page per monitor.
- **Config:** per-part ordered `steps: [{image, description}]`.
- **Content:** per-step images (cropped from the work-standard PDFs).
- **Launcher:** open N Chrome kiosks, one per secondary monitor.
- **Supervisor:** live view of every position in the line.

## Scanner / serial setup
Unchanged from the single-monitor product — one DS3678 feeds the server over a
COM port (optionally shared with the plant's software via a splitter). See
`Serial-Port-Splitter-Setup.md`.

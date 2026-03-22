# Handoff

## State
Phase 1 complete + Phase 2A-C complete (45+ commits, 8 packages passing). Widget Showcase demo at `examples/showcase/`. All controls working: Button, TextView, ImageView, CheckBox, Switch, RadioButton/RadioGroup, ProgressBar, Divider, ScrollView. Text rendering: DirectWrite measurement + GDI ClearType pixels. Anti-aliased shapes via fogleman/gg. DPI-aware.

## Next
1. Phase 2D: EditText (Win32 EDIT control bridge + IME) — last Phase 2 item
2. ScrollView improvements: smooth drag scrolling, scroll indicator, fling tuning
3. Phase 3: RecyclerView, TabLayout, Toolbar, Menu, Dialog

## Context
- `syscall.NewCallback` can't receive float32 params on AMD64 Windows — that's why DWrite DrawGlyphRun callback doesn't work. Using GDI ExtTextOutW for text pixel rendering instead.
- DPI scaling: `scaleNodeTreeDPI` runs once (flag `dpiScaled`), stores `dpiScale` on each node. Widget painters read via `getDPIScale(node)`.
- GGCanvas uses fogleman/gg for shapes but tracks translate offset manually for DrawText/DrawImage (they bypass gg's transform to write directly to raw pixels).
- RadioGroup from XML needs `RegisterButton()` post-inflation (see showcase main.go).
- User is experienced Android dev — use Android naming conventions.

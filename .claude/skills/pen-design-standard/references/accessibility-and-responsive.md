## 5. Accessibility and responsive baseline

**In this file:**

- 5.1.5. Measurable web accessibility baseline
- 5.3.1. Canonical responsive vocabulary


### 5.1.5. Measurable web accessibility baseline

WCAG 2.2 Level AA is the web baseline. The table is a design and implementation review minimum, not an exhaustive WCAG checklist. Measure the rendered implementation as well as the design specification. `CSS px` means CSS pixels, not hardware pixels or screenshot pixels.

| Concern | Required check and relevant distinction |
|---|---|
| Text contrast | At least 4.5:1 for ordinary text and 3:1 for large text: at least 18 pt (24 CSS px), or 14 pt bold (about 18.67 CSS px), with equivalent sizing for CJK fonts. Do not round a failing ratio up. Inactive controls, incidental text, and logotypes have defined exceptions. [SC 1.4.3](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html) |
| Non-text contrast | Visual information needed to identify controls, their states, and meaningful graphics needs 3:1 against adjacent colors, subject to the criterion's exceptions. This is not a blanket contrast requirement for every decorative border. [SC 1.4.11](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html) |
| Text resizing | Text can grow to 200% without loss of content or functionality, except captions and images of text. A fixed-height control must not clip enlarged text. [SC 1.4.4](https://www.w3.org/WAI/WCAG22/Understanding/resize-text.html) |
| Reflow | Vertically scrolling content must work at 320 CSS px width without two-dimensional scrolling; horizontally scrolling content has a corresponding 256 CSS px height condition. Content requiring two-dimensional layout has a limited exception. A 1280 CSS px viewport at 400% zoom is a useful web test. A 390 px design frame alone is insufficient evidence. [SC 1.4.10](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html) |
| Text spacing | In supporting markup, tolerate user overrides of line height to 1.5 times font size, paragraph spacing to 2 times font size, letter spacing to 0.12 em, and word spacing to 0.16 em without losing content or functionality. These are tolerance tests, not mandatory default typography. Account for the criterion's language/script exceptions. [SC 1.4.12](https://www.w3.org/WAI/WCAG22/Understanding/text-spacing.html) |
| Focus | Keyboard focus must be visible. Under AA SC 2.4.11, author-created content must not entirely obscure the focused component; keeping it fully visible is the team preference. Focus Appearance (2.4.13) and Focus Not Obscured (Enhanced, 2.4.12) are AAA, not AA. [Focus guidance](https://www.w3.org/WAI/WCAG22/Understanding/focus-not-obscured-minimum.html) |
| Pointer targets | SC 2.5.8 AA requires at least 24 × 24 CSS px or a qualifying spacing, equivalent-control, inline, user-agent, or essential exception. For the spacing exception, 24 CSS px diameter circles centered on undersized targets must not intersect another target or another such circle. A 44 × 44 target is a useful stronger design choice; do not label it the AA minimum. [SC 2.5.8](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html) |
| Dragging | Provide a single-pointer alternative that does not require dragging unless an applicable exception applies. Keyboard support alone does not satisfy this pointer requirement. [SC 2.5.7](https://www.w3.org/WAI/WCAG22/Understanding/dragging-movements.html) |
| Authentication | Avoid requiring memory, transcription, or puzzle-solving without an allowed alternative, assistance mechanism, or exception. Support password managers and paste; make verification-code entry compatible with paste/autofill where supported. [SC 3.3.8](https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html) |
| Repeated input and help | Reuse or offer previously supplied information within the same process, subject to applicable exceptions. Keep repeated help mechanisms in consistent relative order. [SC 3.3.7](https://www.w3.org/WAI/WCAG22/Understanding/redundant-entry.html), [SC 3.2.6](https://www.w3.org/WAI/WCAG22/Understanding/consistent-help.html) |
| Consequential submissions | For covered legal, financial, and user-data changes, provide reversal, input checking with correction, or review/confirmation before final submission as appropriate. A confirmation dialog is one possible method. [SC 3.3.4](https://www.w3.org/WAI/WCAG22/Understanding/error-prevention-legal-financial-data.html) |

Disabled-state styling should remain understandable under this team standard, even where WCAG contrast exceptions apply. Define the actual disabled behavior; a dimmed appearance alone is not an implementation contract.

Reduced-motion support is a team expectation when motion is used. SC 2.3.3 on disabling interaction-triggered animation is AAA; do not present every reduced-motion recommendation as AA. Other applicable timing, flashing, and moving-content requirements remain part of the baseline. See [Animation from Interactions](https://www.w3.org/WAI/WCAG22/Understanding/animation-from-interactions.html).

---

### 5.3.1. Canonical responsive vocabulary

Regimes describe available layout space, not hardware. Boundaries are set per product in CSS px, non-overlapping, including the exact boundary width.

| Regime | Representative frame width | Example range |
|---|---|---|
| Compact | 390 px | width < 768 CSS px |
| Medium | 768 px | 768 ≤ width < 1200 CSS px |
| Wide | 1440 px | width ≥ 1200 CSS px |

A two-regime product omits Medium; if Medium shares Wide structure, record that mapping rather than drawing a duplicate shell. Representative widths are review frames, not a claim the UI only works at those widths; implementation QA covers just below, at, and above each boundary, plus the reflow condition above.

---

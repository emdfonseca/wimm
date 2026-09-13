## 12. Worked example: create account

### File

Account creation belongs to the onboarding/authentication journey area defined by the same organization used throughout this document:

```text
apps/web/design/journeys/01-onboarding.pen
```

If account creation and authentication are intentionally grouped because the product is still small, use an explicit grouped filename such as:

```text
apps/web/design/journeys/onboarding-and-authentication.pen
```

Do not introduce a generic `product.pen` file solely for this example.

### Imported library

```text
packages/design-system/design/product-ui.lib.pen
```

### Example setup and scope

This example adopts Light/Dark, all three Device values, and three structurally distinct shells: Compact below 768 CSS px, Medium from 768 to below 1200 CSS px, and Wide from 1200 CSS px. Representative frames are 390, 768, and 1440 px wide; choose and record realistic heights. If Medium adds no distinct structure in the actual product, omit that row and document its shared-shell mapping.

Actor: a visitor creating their own account. Account creation and Sign in (`J02`) are distinct journeys even if they share a file. Invitation entry is in scope only with an explicit invitation-validation branch. Copy uses synthetic data; never place credentials or real verification codes in the design file.

### Journey overview

```text
J01 · CREATE ACCOUNT

Goal:
Create a usable account and enter the product.

Entry points:
Marketing site / invitation / direct sign-up route

Primary success outcome:
User reaches initial product home.
```

### Primary success path

```text
J01.A · PRIMARY SUCCESS

                01 Welcome   02 Details   03 Verify   04 Preferences   05 Home

DESKTOP         [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]

TABLET          [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]

MOBILE          [ frame ]    [ frame ]    [ frame ]   [ frame ]        [ frame ]
```

For Mobile frames that use responsive tokens:

```text
Theme:
  Device = Mobile
```

For Tablet and Desktop frames, assign `Device = Tablet` and `Device = Desktop` respectively, and use the mapped Medium and Wide shells.

Main-flow screens can use the product-default color mode:

```text
Theme:
  Color = Light
```

### Failure / recovery paths

```text
J01.B · EXISTING-ACCOUNT HANDLING

02 Details
   ↓
Server detects an existing account (internal condition)
   ↓
Public response follows the reviewed account-disclosure policy
   ↓
Offer Sign in / account recovery without confirming registration status
   OR
Continue through an approved private-channel recovery flow


J01.C · VERIFICATION CODE EXPIRED

03 Verify
   ↓
Expired code
   ↓
Request new code
   ↓
Verification sent
   ↓
Return to Verify


J01.D · VERIFICATION SERVICE ERROR

03 Verify
   ↓
Service unavailable
   ↓
Retry
   OR
Exit and return later
```

Public responses must not automatically reveal that an email address is registered. Review response content, timing, and recovery behavior with security; the server's internal branch can differ from what the interface discloses. See [OWASP Authentication: error messages](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html#authentication-and-error-messages).

### Additional required scenarios for this example

| Scenario | Representation and contract |
|---|---|
| Invalid verification code | Local state at Step 03; preserve entered data appropriately and explain retry. |
| Resend throttled | Local state; explain when retry is available and announce meaningful status without a noisy countdown. |
| Resend or verification times out | Unknown-result handling; check status before duplicating the operation where supported. |
| Invitation invalid, expired, or mismatched | Branch from invitation entry; define request-new-invitation and ordinary sign-up eligibility. |
| User cancels or returns later | Exit/resume branch; define persistence, authentication state, return route, and expired drafts. |
| Verification succeeds on another device | Cross-device handoff; define how the original device discovers completion and what the next step is. |
| Preferences skipped | Alternative success path if optional; required legal agreements must remain distinct from optional preferences. |

Each branch records its rejoin or terminal outcome. These are example-specific acceptance scenarios, not mandatory account architecture for every product.

### Local states near Step 02

```text
02 · DETAILS / MOBILE

[ Default ]
[ Validation error ]
[ Submitting ]
[ Server error ]
```

These are local because the user remains at the same step.

### Theme QA

```text
40 · THEME + ACCESSIBILITY QA

                      LIGHT             DARK
Details form          [ frame ]         [ frame ]
Validation error      [ frame ]         [ frame ]
Verification          [ frame ]         [ frame ]
Success               [ frame ]         [ frame ]
```

Set the right-hand frames to:

```text
Theme → Color = Dark
```

The reusable elements should update through imported semantic variables.

### Accessibility expectations for this journey

```text
Keyboard:
Entire sign-up flow is completable without pointer-only controls.

Authentication:
Support password managers, paste, and accessible verification entry; apply Section 5.1.5.

Validation:
Submitting invalid data identifies each error in text and defines sensible focus behavior.

Verification:
Expired/sent/success statuses require an implementation announcement strategy.

Responsive:
Mobile retains every action required to complete the account.

Themes:
Error, focus, disabled, and success states remain identifiable in Light and Dark.
```

---

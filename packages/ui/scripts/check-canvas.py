#!/usr/bin/env python3
"""Every screen says what its frames say.

The gap this closes: four components were built from their origins and every
screen was composed from a task description instead, so the built screens
diverged from the drawings and nothing failed. check-geometry.py asserts four
measurements on atoms and templates; no screen was checked against any frame.

So: canvas-contract.json holds the static copy of each drawn screen state, and
this asserts the screen that implements those frames contains it.

It checks copy, not layout. A screen can satisfy this and still be arranged
wrongly — but the drift that actually happened was copy and structure ("Your
money" for "Overview", "Refresh" for "Refresh balances", an invented bank row),
and copy is the half a script can hold.

An intentional difference goes in ACCEPTED with a reason. An unexplained one
fails.
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
SRC = Path(__file__).resolve().parents[1] / "src"
CONTRACT = SRC / "canvas-contract.json"

# Which screens implement which frames. A frame prefix maps to every file that
# together renders it: Overview's disconnect confirmation is the screen plus
# the dialog it opens.
IMPLEMENTS = {
    # J01 · Enrol a passkey, J02 · Sign in. The member arrives signed in either
    # way, and that landing is Overview rather than a screen of its own — see
    # ACCEPTED for the placeholder copy the frame still carries from before
    # Overview existed.
    "J01.A / 01 · Enrolment invitation": ["pages/EnrolScreen.svelte"],
    "J01.A / 02 · Creating passkey": ["pages/EnrolScreen.svelte"],
    "J01.A / 03 · Signed in": ["pages/AccountsOverview.svelte"],
    "J01.B / 01 · Link unusable": ["pages/LinkUnusableScreen.svelte"],
    "J02.A / 01 · Sign in": ["pages/SignInScreen.svelte"],
    "J02.A / 02 · Signed in": ["pages/AccountsOverview.svelte"],
    "J02.B / 01 · Passkey not recognised": ["pages/SignInScreen.svelte"],
    # J03 · Connect a bank.
    "J03.A / 01 · Overview": ["pages/AccountsOverview.svelte"],
    "J03.A / 02 · Choose a bank": ["pages/ChooseBankScreen.svelte"],
    "J03.A / 03 · What wimm will see": ["pages/ConsentExplainerScreen.svelte"],
    "J03.A / 04 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    "J06.A / 02 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    # J09 · Hand an account to its owner, leave one out, bring it back.
    "J09.A / 01 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    "J09.A / 02 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    "J09.A / 03 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    "J09.A / 04 · Choose accounts": [
        "pages/ChooseAccountsScreen.svelte",
        "molecules/LeftOutDialog.svelte",
    ],
    "J09.A / 05 · Overview": ["pages/AccountsOverview.svelte", "molecules/AccountRow.svelte"],
    "J09.A / 06 · Choose accounts": ["pages/ChooseAccountsScreen.svelte"],
    # J10 · Name an account and the household.
    "J10.A / 01 · Choose accounts": ["pages/ChooseAccountsScreen.svelte", "molecules/AccountName.svelte"],
    "J10.A / 02 · Overview": ["pages/AccountsOverview.svelte"],
    # J11 · Settings.
    "J11.A / 01 · Settings": ["pages/SettingsScreen.svelte", "molecules/ThemeToggle.svelte"],
    "J11.B / 01 · Settings": [
        "pages/SettingsScreen.svelte",
        "molecules/ThemeToggle.svelte",
        "molecules/DensityControl.svelte",
    ],
    "J03.B / 01 · Overview": ["pages/AccountsOverview.svelte"],
    "J03.C / 01 · Overview": ["pages/AccountsOverview.svelte"],
    "J03.D / 01 · Overview": ["pages/AccountsOverview.svelte"],
    "J04.B / 01 · Overview": ["pages/AccountsOverview.svelte", "molecules/AccountRow.svelte"],
    "J05.A / 01 · Overview": [
        "pages/AccountsOverview.svelte",
        "molecules/AccountRow.svelte",
        "molecules/DisconnectBankDialog.svelte",
    ],
    "J05.A / 02 · Overview": ["pages/AccountsOverview.svelte"],
    "J06.A / 01 · Overview": ["pages/AccountsOverview.svelte", "molecules/AccountRow.svelte"],
    "J06.A / 03 · Overview": ["pages/AccountsOverview.svelte", "molecules/AccountRow.svelte"],
    # J07 · See where the money went. One screen carries every state; the
    # compact frames are the same component with rows stacked, which is why
    # they map to the same file rather than to a second one.
    "J07.A / 00 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.A / 01 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.A / 02 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
    ],
    "J07.A / 03 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.A / 04 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.A / 05 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.B / 01 · Transactions": ["pages/TransactionsScreen.svelte"],
    "J07.B / 02 · Transactions": ["pages/TransactionsScreen.svelte"],
    "J07.B / 03 · Transactions": ["pages/TransactionsScreen.svelte"],
    "J07.B / 04 · Transactions": ["pages/TransactionsScreen.svelte"],
    "J07.C / 01 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.C / 02 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.D / 01 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    "J07.D / 02 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
    # J08 · Include a bank's transactions.
    "J08.A / 01 · What wimm will see": ["pages/WidenConsentScreen.svelte"],
    "J08.A / 02 · Transactions": [
        "pages/TransactionsScreen.svelte",
        "molecules/LedgerRow.svelte",
        "molecules/SeekPager.svelte",
    ],
}

# Copy a screen deliberately does not carry, and why. Each entry is a promise
# that the difference was decided rather than overlooked.
ACCEPTED = {
    # Fixture narration, not template copy. These name the frame's own two
    # members and its own two accounts, so a screen reproducing them literally
    # would say "Grace sees the balance of the joint account" to a household
    # with neither — which is what this check briefly talked me into writing.
    # The screen computes the equivalent sentence from what is actually there.
    # "Alan" is not itself a recognised fixture, so only "Grace" collapses.
    "{member} sees the balance; Alan sees it in full.": "names the frame's fixture members",
    "{member} sees the balance of the joint account. The personal one is nobody's, "
    "and no balance will be read for it.": "names the frame's fixture members",
    # J07.C / 02 draws two banks because the scenario is one answering and one
    # not, so the bank that failed has to be the second one. The screen
    # interpolates whichever bank failed, so the sentence it renders is the
    # drawn one with the right name in it.
    "{bank} did not answer": "the frame's second fixture bank, interpolated in the screen",
    "Everything else is up to date. {bank}'s transactions are the ones last "
    "read at {time}.": "the frame's second fixture bank, interpolated in the screen",
    # J01.A / 03 and J02.A / 02's "Signed in" frames predate Overview: drawn
    # once as a placeholder for "you land here next", before Overview had a
    # design of its own. The member lands on the real Overview, which says
    # what is actually there rather than "nothing here yet".
    "Signed in as {member}": "a placeholder predating Overview; the member lands on Overview itself",
    "There is nothing here yet.": "a placeholder predating Overview; the member lands on Overview itself",
    # J09.A's chooser draws a transient confirmation of the action a member
    # just took — a helper line replacing the standing one, or a notice above
    # the list — naming who now owns or no longer owns an account. Built
    # instead: `invalidateAll()` re-renders the row itself (the other member
    # now shown as an owner, or the account gone from the list and the total),
    # which is the change these sentences narrate. No spec scenario asks for
    # the narration in addition to the visible change, and the local state
    # needed to say "the member you just added" rather than the current
    # owner/grant list is not otherwise part of this screen.
    "{member} owns this account too now. Both of you see it in full, and either "
    "of you can change who else does.": "a transient action confirmation; the row's own change carries it",
    "{account} is {member}'s now": "a transient action confirmation; the row's own change carries it",
    "You stopped being an owner, so it has left your list and your total. "
    "{member} can hand it back.": "a transient action confirmation; the row's own change carries it",
    "Leaving everything else as it is keeps these accounts to yourself. You can "
    "change any of this later.": "the standing helper, not the just-handed-on variant; see above",
    # A page-level "Balances read just now." summary above the account list.
    # Each row already states its own reading (AccountRow's "Reading", a
    # DATA_NODE), which is the freshness fact ADR 0018 requires; an aggregate
    # restating it for the whole page was not built.
    "Balances read just now.": "a page-level freshness summary; each row already states its own reading",
    # J07.A / 03 and 04 draw a "Newer" control beside SeekPager's "Older": both
    # moved to the month scrubber, which reaches every page either did and
    # more (a click on any dot is a page). This is a live redesign, not yet
    # drawn back into the frame — the canvas still shows the pager it
    # replaced.
    "Newer": "superseded by the month scrubber; the frame is not yet redrawn to agree",
}


# Comments are stripped before matching. Every one of these files explains
# itself at length, and the canvas's own wording tends to appear in the
# explanation — so matching raw source lets a screen satisfy this check by
# talking about the copy instead of rendering it.
# The line alternative matches [^\n]* rather than .*: DOTALL applies to the
# whole pattern, so `.*$` ran past every newline and one `//` comment ate the
# rest of the file. That failed loudly here, and it would have passed silently
# in a checker that asserted absence.
COMMENTS = re.compile(
    r"<!--.*?-->"                  # markup
    r"|/\*.*?\*/"                  # block
    r"|^[ \t]*(?://|\*)[^\n]*",    # line, and jsdoc continuation
    re.DOTALL | re.MULTILINE,
)


def rendered(source: str) -> str:
    """The part of a file a member could actually see."""
    return COMMENTS.sub(" ", source)


# A frame's fixture name becomes a placeholder in the contract, and the screen
# interpolates a variable there. Both sides collapse to the same token so the
# sentence around it is what gets compared.
INTERPOLATION = re.compile(r"\{bank\}|\{member\}|\{[A-Za-z?.()\[\]'\" ]+\}")


def shape(text: str) -> str:
    """The sentence with whatever fills its holes reduced to one token."""
    return INTERPOLATION.sub("\u2022", text)


def normalise(text: str) -> str:
    """Curly quotes and dashes differ between a canvas and a source file."""
    for fancy, plain in (("’", "'"), ("‘", "'"), ("“", '"'),
                         ("”", '"'), ("—", "-"), ("–", "-")):
        text = text.replace(fancy, plain)
    return " ".join(text.split())


def check(contract, src, implements, accepted):
    """Every problem with these screens, and how many strings were compared.

    Takes its inputs rather than reading the module's own constants, so the
    check can be fed the shapes that must fail. Asserting that the current
    tree passes proves nothing about the check: the current tree is valid.
    """
    sources = {}
    problems = []
    checked = 0

    for frame, lines in sorted(contract.items()):
        files = next((f for prefix, f in implements.items() if frame.startswith(prefix)), None)
        if files is None:
            problems.append(
                f"{frame}: no screen is recorded as implementing this frame. "
                f"Add it to IMPLEMENTS in {Path(__file__).name}, or the frame is drawn and unbuilt."
            )
            continue

        haystack = ""
        for name in files:
            if name not in sources:
                sources[name] = shape(normalise(rendered((src / name).read_text())))
            haystack += sources[name]

        for line in lines:
            wanted = shape(normalise(line))
            checked += 1
            if wanted in haystack:
                continue
            if line in accepted:
                continue
            problems.append(
                f"{frame}: the canvas says {line!r}\n"
                f"      and {', '.join(files)} does not. Match the drawing, or record the "
                f"difference in ACCEPTED with a reason."
            )

    return problems, checked


def main() -> int:
    contract = json.loads(CONTRACT.read_text())
    problems, checked = check(contract, SRC, IMPLEMENTS, ACCEPTED)

    if problems:
        print(f"{len(problems)} screen(s) diverge from the canvas:\n", file=sys.stderr)
        for problem in problems:
            print(f"  - {problem}", file=sys.stderr)
        return 1

    print(f"screens match the canvas ({checked} strings across {len(contract)} frames)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

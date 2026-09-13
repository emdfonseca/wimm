# What gets a story

## The mapping

| Design layer | Storybook | Notes |
|---|---|---|
| Foundations | `Foundations/*` MDX docs | Read live CSS custom properties; no story per token |
| Atom | `Atoms/<Name>` | Every variant an arg, every state a story |
| Molecule | `Molecules/<Name>` | Includes its composed states (field + error + hint) |
| Organism | `Organisms/<Name>` | Mock data as args; no network |
| Template | `Templates/<Name>` | Slots filled with placeholder regions, not real pages |
| Page | `Pages/<Name>` | The **pure** screen component: data as props, actions as callbacks |
| (route module) | — | `+page.svelte` / `+page.server.ts` wire data and navigation. Not a story. |
| Journey | — | Lives in the `.pen` file; verified by e2e tests |

## Pages: split pure from connected

Sign-in, dashboard, and checkout review are exactly the screens most worth reviewing in isolation — their empty, error, and loading states are where design and implementation usually disagree. They get stories. What does not get a story is the *connected* half.

```text
src/routes/sign-in/
├── +page.server.ts          load + form actions          ← no story
└── +page.svelte             reads data, calls goto       ← no story, stays thin

src/lib/screens/SignIn/
├── SignIn.svelte            props in, callbacks out      ← Pages/SignIn
└── SignIn.stories.svelte
```

The route module's only job is to fetch, wire, and navigate. Everything visual lives in the screen component, which receives data as props and emits intent as callbacks (`onsubmit`, `onretry`). That split is worth doing for its own sake — it makes the screen reviewable, testable, and reusable across entry points — and the story falls out of it for free.

**The mocking burden is the test.** If a Page story needs `$app/state`, a load function, or `fetch` mocked to render, the screen is still connected and the fix is decomposition, not `sveltekit_experimental`. A little navigation mocking for anchors is fine; mocking a data layer is the signal.

## Why journeys still stop at the boundary

A journey is sign-in *through* error, recovery, and landing on the dashboard — routing, session, and real navigation across screens. Reproducing that here means mocking the router, load functions, session, and API, and what you end up verifying is the quality of the mocks. It looks like coverage and is not. Journeys live in the `.pen` file and are verified by e2e tests against the running app.

So the line is not "Pages are too high-level". It is: **Storybook renders a screen in a state; the app is what proves a flow.**

## Templates versus Pages

A Template is the reusable scaffold with placeholder regions — it belongs to the design system, and its story proves the scaffold responds. A Page is that scaffold populated with a specific product screen's real structure and content, and it belongs to the product. Both get stories; they answer different questions, which is why the placeholder rule below applies only to Templates.

## Templates need placeholder content, not real content

```svelte
<Story name="Wide">
  {#snippet template(args)}
    <AppShell {...args}>
      {#snippet nav()}<PlaceholderNav />{/snippet}
      {#snippet main()}<PlaceholderBlock height="60vh" label="Page content" />{/snippet}
    </AppShell>
  {/snippet}
</Story>
```

Placeholder regions keep the story about the scaffold. Real content makes the story about the content, and it will be edited for the wrong reasons forever after.

## App Shells

The design system owns a canonical shell per regime (Compact / Medium / Wide). Those are Templates, and each gets a story — but the regimes are viewport globals, not three separate stories, unless the shells are genuinely different compositions rather than one fluid shell. Match whatever the `.lib.pen` library actually defines: if it has one shell that adapts, that is one story viewed at three viewports; if it has three compositions, that is three stories.

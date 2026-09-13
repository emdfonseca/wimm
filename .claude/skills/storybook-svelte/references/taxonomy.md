# What gets a story

## The mapping

| Design layer | Storybook | Notes |
|---|---|---|
| Foundations | `Foundations/*` MDX docs | Read live CSS custom properties; no story per token |
| Atom | `Atoms/<Name>` | Every variant an arg, every state a story |
| Molecule | `Molecules/<Name>` | Includes its composed states (field + error + hint) |
| Organism | `Organisms/<Name>` | Mock data as args; no network |
| Template | `Templates/<Name>` | Slots filled with placeholder regions, not real pages |
| Page | — | A SvelteKit route. Not a story. |
| Journey | — | Lives in the `.pen` file; verified by e2e tests |

## Why Pages stop at the boundary

A Page is a Template populated with real content, real data, real routing, and real navigation state. Reproducing that in Storybook means mocking the router, the load functions, the session, and the API — and what you end up verifying is the quality of your mocks. It is slow to run, expensive to maintain, and it quietly makes people believe flows are covered when they are not.

Templates are where Storybook stops being useful for layout: a Template story with placeholder regions proves the scaffold responds correctly, which is exactly what a Template promises. Everything above it is a flow, and flows are verified by running the app.

When someone asks for a "page story", the useful response is to identify which Template and Organisms it is made of, story those, and cover the flow with an e2e test.

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

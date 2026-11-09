# Theme
## Part 1 — Token summary
Dark-only UI. Tailwind v4 default palette, mostly `slate-*`, with a few `gray-*`.
- CSS vars (`src/routes/layout.css` :root): --background #0b0f19, --surface #111827, --surface-hover #1f2937, --border #273244, --text-primary #f9fafb, --text-secondary #9ca3af, --accent #3b82f6, --success #22c55e, --danger #ef4444
- Body bg: radial-gradient(circle at top, #172554 0%, #0b0f19 35%); main overrides with bg-slate-950
- Surfaces: cards bg-slate-900, border-slate-800, hover bg-slate-800; table header bg-slate-950/50
- Text: white / slate-400 secondary / slate-500 muted
- Accent: blue-600 buttons (hover blue-500), links blue-400
- Semantic: positive emerald-400, negative red-400; badges use light *-100/*-700 pills (inconsistent with dark theme)
- Font: Inter, system-ui, sans-serif (Inter not loaded explicitly)
- Radius: rounded-xl buttons, rounded-2xl cards, rounded-full avatars/badges
- Shadows: shadow-xl on SummaryCard, shadow-lg shadow-blue-600/20 on brand mark
- Spacing: cards p-5/p-6, gaps 4; breakpoints Tailwind defaults (sm 640, md 768, lg 1024); max-w-7xl container
- No tailwind.config (v4 CSS-first); no dark-mode toggle

## Part 2 — Raw source

### `src/routes/layout.css`

```css
@import "tailwindcss";


:root {
    --background: #0b0f19;
    --surface: #111827;
    --surface-hover: #1f2937;

    --border: #273244;

    --text-primary: #f9fafb;
    --text-secondary: #9ca3af;

    --accent: #3b82f6;
    --success: #22c55e;
    --danger: #ef4444;
}


html {
    background: var(--background);
    color: var(--text-primary);
}


body {
    margin:0;
    min-height: 100vh;
    background:
            radial-gradient(
                    circle at top,
                    #172554 0%,
                    var(--background) 35%
            );

    color: var(--text-primary);
    font-family:
            Inter,
            system-ui,
            sans-serif;
}


/* scrollbar */

::-webkit-scrollbar {
    width: 8px;
}


::-webkit-scrollbar-track {
    background: var(--background);
}


::-webkit-scrollbar-thumb {
    background: #374151;
    border-radius: 999px;
}


::-webkit-scrollbar-thumb:hover {
    background: #4b5563;
}
```

### `vite.config.ts`

```ts
import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-auto';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
			// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
			// See https://svelte.dev/docs/kit/adapters for more information about adapters.
			adapter: adapter()
		})
	]
});

```

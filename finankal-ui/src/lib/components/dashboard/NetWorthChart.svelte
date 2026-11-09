<script lang="ts">
    import { TrendingUp } from "lucide-svelte";

    type Point = { label: string; value: number };

    let { points = [] }: { points: Point[] } = $props();

    const W = 600;
    const H = 220;
    const PAD = 20;

    // Map values into the SVG box; flat series sits mid-height.
    const coords = $derived.by(() => {
        if (points.length < 2) return [];
        const values = points.map((p) => p.value);
        const min = Math.min(...values);
        const range = Math.max(...values) - min || 1;
        return points.map((p, i) => [
            (i / (points.length - 1)) * W,
            H - PAD - ((p.value - min) / range) * (H - PAD * 2)
        ]);
    });

    const line = $derived(coords.map(([x, y]) => `${x},${y}`).join(" "));
    const area = $derived(coords.length ? `M0,${H} L${line.replaceAll(" ", " L")} L${W},${H} Z` : "");
</script>

<div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
    <div class="flex items-center justify-between">
        <div>
            <h2 class="text-lg font-semibold text-white">Net worth over time</h2>
            <p class="mt-1 text-sm text-slate-400">Last {points.length || 6} months</p>
        </div>
        <div class="flex h-10 w-10 items-center justify-center rounded-xl bg-slate-800 text-blue-400">
            <TrendingUp class="h-5 w-5" />
        </div>
    </div>

    {#if coords.length === 0}
        <div class="mt-6 flex h-56 items-center justify-center rounded-xl border border-dashed border-slate-800 text-sm text-slate-500">
            Not enough history yet to chart your net worth.
        </div>
    {:else}
        <div class="mt-6">
            <svg viewBox="0 0 {W} {H}" class="h-56 w-full" preserveAspectRatio="none" role="img" aria-label="Net worth trend">
                <defs>
                    <linearGradient id="netWorthFill" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.25" />
                        <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
                    </linearGradient>
                </defs>
                {#each [20, 70, 120, 170] as y}
                    <line x1="0" y1={y} x2={W} y2={y} stroke="#1e293b" stroke-width="1" />
                {/each}
                <path d={area} fill="url(#netWorthFill)" />
                <polyline points={line} fill="none" stroke="#3b82f6" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" vector-effect="non-scaling-stroke" />
            </svg>
            <div class="mt-2 flex justify-between text-xs text-slate-500">
                {#each points as p}<span>{p.label}</span>{/each}
            </div>
        </div>
    {/if}
</div>

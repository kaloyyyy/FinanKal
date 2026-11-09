<script lang="ts">
    import { currency } from "$lib/utils/currency";
    import type { PositionTile } from "$lib/utils/payables";

    let { tile, large = false }: { tile: PositionTile; large?: boolean } = $props();
</script>

<div class="flex h-full flex-col rounded-xl border border-slate-800 bg-slate-950/50 p-5">
    <p class="text-xs font-semibold tracking-wider text-slate-400 uppercase">{tile.label}</p>
    <p class={`mt-2 font-bold tabular-nums ${large ? "text-4xl tracking-tight" : "text-2xl"} ${tile.value < 0 ? "text-rose-400" : "text-emerald-400"}`}>
        {currency.format(tile.value)}
    </p>
    <p class="mt-1 text-xs text-slate-500">{tile.hint}</p>

    <dl class="mt-4 space-y-1.5 border-t border-slate-800 pt-3 text-sm">
        {#each tile.terms as term}
            <div class="flex justify-between gap-3">
                <dt class="text-slate-400">
                    <span class="inline-block w-3 text-slate-500" aria-hidden="true">{term.sign}</span>
                    <span class="sr-only">{term.sign === "−" ? "minus" : term.sign === "+" ? "plus" : ""}</span>
                    {term.label}
                </dt>
                <dd class="text-slate-200 tabular-nums">{currency.format(term.amount)}</dd>
            </div>
        {/each}
    </dl>
</div>

<script lang="ts">
    import { currency } from "$lib/utils/currency";
    import { formatDate, formatMonth } from "$lib/utils/date";
    import { positionTiles, type NetPositions, type PayableMonth } from "$lib/utils/payables";
    import NetPositionTile from "./NetPositionTile.svelte";

    let {
        thisMonth,
        nextMonth,
        positions
    }: { thisMonth: PayableMonth; nextMonth: PayableMonth; positions: NetPositions } = $props();

    const month = (key: string) => formatMonth(`${key}-01`);

    const tiles = $derived(positionTiles(positions, thisMonth.month, nextMonth.month, month));

    const dues = $derived([
        { label: "Due this month", ...thisMonth },
        { label: "Due next month", ...nextMonth }
    ]);
</script>

<section class="rounded-2xl border border-slate-800 bg-slate-900" aria-labelledby="positions-heading">
    <div class="p-6 pb-2">
        <h2 id="positions-heading" class="text-lg font-semibold text-white">Net position</h2>
        <p class="mt-1 text-sm text-slate-400">What you have against what you owe, now and for the next two months.</p>
    </div>

    <div class="grid grid-cols-1 gap-4 p-6 pt-4 md:grid-cols-3">
        {#each tiles as tile}
            <NetPositionTile {tile} />
        {/each}

    </div>

    <div class="grid grid-cols-1 border-t border-slate-800 sm:grid-cols-2 sm:divide-x sm:divide-slate-800">
        {#each dues as due}
            <div class="border-b border-slate-800 p-6 last:border-b-0 sm:border-b-0">
                <div class="flex items-baseline justify-between gap-3">
                    <p class="text-sm font-medium text-slate-300">
                        {due.label} <span class="text-slate-500">· {month(due.month)}</span>
                    </p>
                    <p class={`text-base font-semibold tabular-nums ${due.total > 0 ? "text-amber-400" : "text-emerald-400"}`}>
                        {currency.format(due.total)}
                    </p>
                </div>

                {#if due.items.length}
                    <ul class="mt-3 space-y-2">
                        {#each due.items as item}
                            <li class="flex items-start justify-between gap-3 text-sm">
                                <span class="min-w-0">
                                    <span class="block truncate text-slate-200">{item.name}</span>
                                    <span class="text-xs text-slate-500 tabular-nums">
                                        Due {formatDate(item.dueDate)}{item.installmentNumber && item.installments ? ` · ${item.installmentNumber}/${item.installments}` : ""}
                                    </span>
                                </span>
                                <span class="shrink-0 text-slate-300 tabular-nums">{currency.format(item.amount)}</span>
                            </li>
                        {/each}
                    </ul>
                {:else}
                    <p class="mt-3 text-sm text-slate-500">Nothing left to pay.</p>
                {/if}
            </div>
        {/each}
    </div>
</section>

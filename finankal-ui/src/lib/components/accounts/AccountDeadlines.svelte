<script lang="ts">
    import { currency } from "$lib/utils/currency";
    import { formatDate } from "$lib/utils/date";
    import type { Payable } from "$lib/utils/payables";

    let { deadlines }: { deadlines: Payable[] } = $props();

    const SHOWN = 3; // upcoming rows per plan before "show more"

    // Days from today to a due date, on the Philippine calendar.
    const today = new Date(Date.now() + 8 * 60 * 60 * 1000).toISOString().slice(0, 10);
    const daysUntil = (due: string) => Math.round((Date.parse(due) - Date.parse(today)) / 86_400_000);
    const when = (days: number) =>
        days < 0 ? `${-days} ${days === -1 ? "day" : "days"} overdue` : days === 0 ? "due today" : `in ${days} ${days === 1 ? "day" : "days"}`;
    const urgency = (days: number) => (days < 0 ? "text-red-400" : days <= 7 ? "text-amber-300" : "text-slate-400");

    // Deadlines grouped by the schedule (or card bill) they come from, soonest plan first.
    const plans = $derived.by(() => {
        const byName = new Map<string, Payable[]>();
        for (const d of deadlines) byName.set(d.name, [...(byName.get(d.name) ?? []), d]);
        return [...byName].map(([name, items]) => {
            const term = items[0].installments; // the loan's last installment number
            // Installments before the first unpaid one are settled; card bills aren't numbered.
            const settled = items[0].installmentNumber ? items[0].installmentNumber - 1 : null;
            const left = items.reduce((cents, d) => cents + Math.round(Number(d.amount) * 100), 0) / 100;
            return { name, items, term, settled, left };
        });
    });

    const total = $derived(plans.reduce((cents, p) => cents + Math.round(p.left * 100), 0) / 100);
    const next = $derived(deadlines[0]);
</script>

<div class="space-y-4">
    <div class="flex flex-wrap items-end justify-between gap-x-6 gap-y-2">
        {#if next}
            <p class="text-sm">
                <span class="text-slate-400">Next</span>
                <span class="font-semibold text-white tabular-nums">{currency.format(Number(next.amount))}</span>
                <span class={urgency(daysUntil(next.dueDate))}>{when(daysUntil(next.dueDate))}</span>
                <span class="text-slate-400">on {formatDate(next.dueDate)}</span>
            </p>
        {/if}
        <p class="text-sm text-slate-400">
            Still to pay <span class="font-semibold text-white tabular-nums">{currency.format(total)}</span>
        </p>
    </div>

    {#each plans as plan (plan.name)}
        <article class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
                <h3 class="min-w-0 font-medium text-white">{plan.name}</h3>
                <p class="text-sm text-slate-400 tabular-nums">
                    {currency.format(plan.left)} left
                </p>
            </div>

            {#if plan.term && plan.settled !== null}
                <!-- One segment per installment: settled ones filled, the next one marked. -->
                <div class="mt-4 flex gap-[3px]" role="img"
                     aria-label={`${plan.settled} of ${plan.term} installments paid`}>
                    {#each { length: plan.term } as _, i}
                        <span class={`h-2 flex-1 rounded-[2px] ${
                            i < plan.settled ? "bg-blue-500"
                            : i === plan.settled ? "bg-blue-500/35 ring-1 ring-blue-400 ring-inset"
                            : "bg-slate-800"}`}></span>
                    {/each}
                </div>
                <p class="mt-2 text-xs text-slate-500 tabular-nums">
                    {plan.settled} of {plan.term} paid
                </p>
            {/if}

            <ul class="mt-4 divide-y divide-slate-800 border-t border-slate-800">
                {#each plan.items.slice(0, SHOWN) as d (d.dueDate)}
                    {@render row(d)}
                {/each}
            </ul>
            {#if plan.items.length > SHOWN}
                <details class="group">
                    <summary class="mt-1 cursor-pointer list-none rounded-lg py-1.5 text-sm font-medium text-blue-400 hover:text-blue-300 focus-visible:outline-2 focus-visible:outline-blue-500">
                        <span class="group-open:hidden">Show {plan.items.length - SHOWN} more</span>
                        <span class="hidden group-open:inline">Show fewer</span>
                    </summary>
                    <ul class="divide-y divide-slate-800 border-t border-slate-800">
                        {#each plan.items.slice(SHOWN) as d (d.dueDate)}
                            {@render row(d)}
                        {/each}
                    </ul>
                </details>
            {/if}
        </article>
    {/each}
</div>

{#snippet row(d: Payable)}
    {@const days = daysUntil(d.dueDate)}
    <li class="flex items-center justify-between gap-4 py-2.5 text-sm">
        <span class="flex min-w-0 items-baseline gap-3">
            {#if d.installmentNumber}
                <span class="w-6 shrink-0 text-right text-xs text-slate-500 tabular-nums">{d.installmentNumber}</span>
            {/if}
            <span class="font-medium text-slate-200 tabular-nums">{formatDate(d.dueDate)}</span>
            {#if days <= 7}
                <span class={`text-xs ${urgency(days)}`}>{when(days)}</span>
            {/if}
        </span>
        <span class="shrink-0 font-semibold text-white tabular-nums">{currency.format(Number(d.amount))}</span>
    </li>
{/snippet}

<script lang="ts">
    import { onMount } from "svelte";
    import { currency } from "$lib/utils/currency";
    import { formatDate, formatMonth } from "$lib/utils/date";
    import { positionTiles, type NetPositions, type Payable, type PayableMonth } from "$lib/utils/payables";
    import type { UnionBankInstallmentPlan } from "$lib/utils/unionBankInstallments";
    import type { Account } from "$lib/server/api/accounts";
    import NetPositionTile from "./NetPositionTile.svelte";
    import PaikotCard from "./PaikotCard.svelte";

    let {
        thisMonth,
        nextMonth,
        positions,
        paikot,
        paikotMoves,
        paikotError = "",
        accounts,
        userId
    }: { thisMonth: PayableMonth; nextMonth: PayableMonth; positions: NetPositions; paikot: PayableMonth; paikotMoves: Payable[]; paikotError?: string; accounts: Account[]; userId: string } = $props();

    const month = (key: string) => formatMonth(`${key}-01`);
    const plansKey = `finankal-card-installments-v1-${userId}`;
    let cardPlans = $state<UnionBankInstallmentPlan[]>([]);
    let plansReady = $state(false);
    const nextInstallments = $derived.by(() => cardPlans.flatMap((plan) => {
        const card = accounts.find((account) => account.id === plan.cardId);
        if (!card) return [];
        return plan.rows.filter((row) => row.number >= plan.nextNumber && row.number > 0).flatMap((row) => {
            const offset = row.number - plan.nextNumber;
            const [year, month, day] = plan.nextDueDate.split("-").map(Number);
            const targetMonth = month - 1 + offset;
            const endDay = new Date(Date.UTC(year, targetMonth + 1, 0)).getUTCDate();
            const dueDate = new Date(Date.UTC(year, targetMonth, Math.min(day, endDay))).toISOString().slice(0, 10);
            return dueDate.slice(0, 7) === nextMonth.month ? [{
                name: plan.name, accountId: card.id, accountName: card.name, dueDate,
                amount: row.principal + row.interest, installmentNumber: row.number, installments: plan.term,
                principal: row.principal, interest: row.interest
            } satisfies Payable & { principal: number; interest: number }] : [];
        });
    }));
    const dueNextMonth = $derived({ ...nextMonth, total: nextMonth.total + nextInstallments.reduce((sum, item) => sum + item.amount, 0), items: [...nextMonth.items, ...nextInstallments].sort((a, b) => a.dueDate.localeCompare(b.dueDate)) });
    const forecastPositions = $derived({ ...positions, dueNextMonth: positions.dueNextMonth + nextInstallments.reduce((sum, item) => sum + item.amount, 0), nextMonth: positions.money + positions.salary - positions.dueNextMonth - nextInstallments.reduce((sum, item) => sum + item.amount, 0) });
    const tiles = $derived(positionTiles(forecastPositions, thisMonth.month, nextMonth.month, month));
    const dues = $derived([
        { label: "Due this month", ...thisMonth },
        { label: "Due next month", ...dueNextMonth }
    ]);

    onMount(() => {
        try {
            const saved: unknown = JSON.parse(localStorage.getItem(plansKey) ?? "[]");
            if (Array.isArray(saved)) cardPlans = saved.filter((plan): plan is UnionBankInstallmentPlan =>
                !!plan && typeof plan === "object" && typeof (plan as UnionBankInstallmentPlan).id === "string" &&
                typeof (plan as UnionBankInstallmentPlan).cardId === "string" && Array.isArray((plan as UnionBankInstallmentPlan).rows));
        } catch { cardPlans = []; }
        plansReady = true;
    });

    $effect(() => {
        if (plansReady) {
            try { localStorage.setItem(plansKey, JSON.stringify(cardPlans)); } catch { /* forecast remains available until this page closes */ }
        }
    });

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

    <div class="grid grid-cols-1 border-t border-slate-800 sm:grid-cols-3 sm:divide-x sm:divide-slate-800">
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
                                        Due {formatDate(item.dueDate)}{item.installmentNumber && item.installments ? ` · ${item.installmentNumber}/${item.installments}` : ""}{"principal" in item ? ` · principal ${currency.format(item.principal)}, interest ${currency.format(item.interest)}` : ""}
                                    </span>
                                </span>
                                <span class="shrink-0 text-slate-300 tabular-nums">{currency.format(item.amount)}</span>
                            </li>
                        {/each}
                    </ul>
                {:else}
                    <p class="mt-3 text-sm text-slate-500">Nothing left to pay.</p>
                {/if}

                {#if due.label === "Due next month" && cardPlans.length}
                    <div class="mt-4 space-y-2 border-t border-slate-800 pt-3">
                        <p class="text-xs font-medium text-slate-300">Credit card installment forecasts</p>
                        {#each cardPlans as plan}
                            <div class="flex items-start justify-between gap-2 rounded-lg bg-slate-950/70 p-2 text-xs">
                                <span class="min-w-0"><span class="block truncate text-slate-300">{plan.name} · {plan.reference || "No reference"}</span><span class="text-slate-500">{currency.format(plan.monthlyInstallment)} monthly · next #{plan.nextNumber} {formatDate(plan.nextDueDate)}</span></span>
                                <button class="text-slate-500 hover:text-rose-300" aria-label="Remove installment forecast" onclick={() => cardPlans = cardPlans.filter((entry) => entry.id !== plan.id)}>Remove</button>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        {/each}
        <PaikotCard {paikot} plain={nextMonth} moves={paikotMoves} {accounts} error={paikotError} />
    </div>
</section>

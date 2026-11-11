<script lang="ts">
    import { enhance } from "$app/forms";
    import { currency } from "$lib/utils/currency";
    import { formatDate } from "$lib/utils/date";
    import type { Payable } from "$lib/utils/payables";

    let { deadlines, isLoan = false, isCard = false, isMayaLoan = false, payFrom = [], error = "" }: { deadlines: Payable[]; isLoan?: boolean; isCard?: boolean; isMayaLoan?: boolean; payFrom?: { id: string; name: string }[]; error?: string } = $props();
    let paying = $state(false);

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
        for (const d of deadlines) { const key=d.scheduleId ?? d.name; byName.set(key, [...(byName.get(key) ?? []), d]); }
        return [...byName].map(([key, items]) => {
            const name=items[0].name;
            const term = items[0].installments; // the loan's last installment number
            // Installments before the first unpaid one are settled; card bills aren't numbered.
            const settled = items[0].installmentNumber ? items[0].installmentNumber - 1 : null;
            const left = items.reduce((cents, d) => cents + Math.round(Number(d.amount) * 100), 0) / 100;
            return { key, name, items, term, settled, left };
        });
    });

    const total = $derived(plans.reduce((cents, p) => cents + Math.round(p.left * 100), 0) / 100);
    const next = $derived(deadlines[0]);
    const canPay = (d: Payable) => (isLoan || isCard) && !!d.scheduleId && !!d.occurrenceId && !!d.sourceAccountId && deadlines.find((item) => item.scheduleId === d.scheduleId) === d;
</script>

<div class="space-y-4">
    {#if error}
        <p role="alert" class="rounded-xl border border-red-500/20 bg-red-500/10 px-3 py-2 text-sm text-red-300">{error}</p>
    {/if}
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

    {#each plans as plan (plan.key)}
        <article class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
                <h3 class="min-w-0 font-medium text-white">{plan.name}</h3>
                <p class="text-sm text-slate-400 tabular-nums">
                    {currency.format(plan.left)} left
                </p>
            </div>

            {#if plan.items[0].scheduleId}
                <details class="mt-3 rounded-lg border border-slate-800 p-3">
                    <summary class="cursor-pointer text-sm text-slate-300">Plan settings</summary>
                    <form method="POST" action="?/updateInstallments" class="mt-3 space-y-3">
                        <input type="hidden" name="scheduleId" value={plan.items[0].scheduleId} />
                        <input type="hidden" name="firstInstallmentNumber" value={plan.items[0].installmentNumber ?? 1} />
                        <label class="block text-xs text-slate-400">Plan name<input name="name" required value={plan.name} class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white" /></label>
                        <label class="block text-xs text-slate-400">Pay from<select name="sourceAccountId" required value={plan.items[0].sourceAccountId} class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white">{#each payFrom as account}<option value={account.id}>{account.name}</option>{/each}</select></label>
                        <label class="block text-xs text-slate-400">Pending installments · one date and amount per line<textarea name="rows" rows={Math.min(5,Math.max(3,plan.items.length))} required class="mt-1 w-full max-h-36 rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-xs text-white">{plan.items.map((item) => `${item.scheduledDate ?? item.dueDate}, ${item.amount}`).join("\n")}</textarea></label>
                        <button class="rounded-lg bg-blue-600 px-3 py-2 text-xs font-semibold text-white">Save pending schedule</button>
                    </form>
                    {#if isLoan}
                        {#if plan.key === plans[0]?.key}
                            <form method="POST" action="?/deleteLoan" class="mt-3 border-t border-slate-800 pt-3" onsubmit={(e) => !confirm("Delete this loan and all its installment plans? The loan calculation will be removed; the account and ledger history will be kept.") && e.preventDefault()}>
                                <button class="text-xs text-red-300 hover:text-red-200">Delete loan and all plans</button>
                            </form>
                        {/if}
                    {:else}
                        <form method="POST" action="?/deleteInstallments" class="mt-3 border-t border-slate-800 pt-3" onsubmit={(e) => !confirm("Cancel the remaining installments in this plan?") && e.preventDefault()}>
                            <input type="hidden" name="scheduleId" value={plan.items[0].scheduleId} />
                            <button class="text-xs text-red-300 hover:text-red-200">Delete installment plan</button>
                        </form>
                    {/if}
                </details>
            {/if}

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
        <span class="flex shrink-0 flex-col items-end gap-1">
            <span class="font-semibold text-white tabular-nums">{currency.format(Number(d.amount))}</span>
            {#if canPay(d)}
                <form method="POST" action="?/payInstallment" class={`flex flex-col items-end gap-1 ${isMayaLoan ? "rounded-lg border border-slate-800 bg-slate-950/40 p-2" : ""}`} use:enhance={({ formElement }) => {
                    const checkbox = formElement.querySelector<HTMLInputElement>('input[type="checkbox"]');
                    paying = true;
                    return async ({ update }) => {
                        paying = false;
                        if (checkbox) checkbox.checked = false;
                        await update({ reset: false });
                    };
                }}>
                    <input type="hidden" name="name" value={d.name} />
                    <input type="hidden" name="dueDate" value={d.dueDate} />
                    <input type="hidden" name="installmentNumber" value={d.installmentNumber ?? ""} />
                    <input type="hidden" name="sourceAccountId" value={d.sourceAccountId} />
                    <input type="hidden" name="scheduleId" value={d.scheduleId} />
                    <input type="hidden" name="occurrenceId" value={d.occurrenceId} />
                    {#if isMayaLoan}
                        <label class="flex items-center gap-2 text-xs text-slate-300">
                            Amount to pay
                            <span class="relative"><span class="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-slate-500">₱</span><input name="paymentAmount" type="number" inputmode="decimal" min={d.amount} step="0.01" value={d.amount} class="w-36 rounded-md border border-slate-700 bg-slate-950 py-1.5 pl-6 pr-2 text-right font-semibold tabular-nums text-white" aria-label={`Payment amount for installment ${d.installmentNumber ?? ""}`} /></span>
                        </label>
                        <p class="max-w-64 text-right text-xs text-slate-400">Pay extra to reduce principal and advance future installments.</p>
                    {/if}
                    <label class="inline-flex cursor-pointer items-center gap-1.5 text-xs text-emerald-300 has-[:disabled]:cursor-wait has-[:disabled]:opacity-50">
                        <input type="checkbox" disabled={paying} aria-label="Record payment from saved bank or wallet" onchange={(event) => {
                            if (event.currentTarget.checked) event.currentTarget.form?.requestSubmit();
                        }} class="h-4 w-4 rounded border-slate-600 bg-slate-950 text-emerald-500 focus:ring-emerald-500" />
                        {paying ? "Recording…" : `${isMayaLoan ? "Record payment" : "Mark paid"} · from ${payFrom.find((account) => account.id === d.sourceAccountId)?.name ?? "saved account"}`}
                    </label>
                </form>
            {/if}
        </span>
    </li>
{/snippet}

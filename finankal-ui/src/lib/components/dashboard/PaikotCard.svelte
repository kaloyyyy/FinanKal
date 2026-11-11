<script lang="ts">
    import { enhance } from "$app/forms";
    import { currency } from "$lib/utils/currency";
    import { formatDate, formatMonth } from "$lib/utils/date";
    import type { Payable, PayableMonth } from "$lib/utils/payables";
    import type { Account } from "$lib/server/api/accounts";

    let {
        paikot,
        plain,
        moves,
        accounts,
        error = ""
    }: {
        paikot: PayableMonth; // next month's dues with every linked due moved onto its card
        plain: PayableMonth; // next month's dues as they are
        moves: Payable[]; // every due a link moves
        accounts: Account[];
        error?: string;
    } = $props();

    const cents = (n: number) => Math.round(Number(n) * 100);
    const field = "mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white";

    // A card needs a billing day to have a cycle a due can move into.
    const cards = $derived(accounts.filter((a) => a.type === "CREDIT_CARD" && a.billingDay));
    const debts = $derived(accounts.filter((a) => a.type === "LIABILITY" || a.type === "CREDIT_CARD"));
    // One link per account whose dues move.
    const links = $derived([...new Map(moves.map((m) => [m.accountId, m])).values()]);
    // The moves that change next month: dues from now to the end of it.
    const soon = $derived(moves.filter((m) => m.dueDate.slice(0, 7) <= paikot.month));
    // What those moves charge each card, against the credit it has left today.
    const charges = $derived(
        cards
            .map((card) => ({
                card,
                charged: soon.filter((m) => m.paikot!.cardId === card.id).reduce((c, m) => c + cents(m.amount) + cents(m.paikot!.fee), 0) / 100,
                available: (cents(card.creditLimit ?? 0) + cents(card.balance)) / 100
            }))
            .filter((c) => c.charged > 0)
    );
    const less = $derived((cents(plain.total) - cents(paikot.total)) / 100);
    const payWindow = (m: Payable) =>
        m.paikot!.payFrom === m.paikot!.payBy ? `on ${formatDate(m.paikot!.payBy)}` : `${formatDate(m.paikot!.payFrom)} – ${formatDate(m.paikot!.payBy)}`;
</script>

<div class="p-6">
    <div class="flex items-baseline justify-between gap-3">
        <p class="text-sm font-medium text-slate-300">
            Paikot <span class="text-slate-500">· {formatMonth(`${paikot.month}-01`)}</span>
        </p>
        <p class={`text-base font-semibold tabular-nums ${paikot.total > 0 ? "text-amber-400" : "text-emerald-400"}`}>
            {currency.format(paikot.total)}
        </p>
    </div>
    <p class="mt-1 text-xs text-slate-500">
        Next month's dues once each linked card pays its debts.
        {#if less}{currency.format(Math.abs(less))} {less > 0 ? "less" : "more"} than without it.{/if}
    </p>

    {#if soon.length}
        <ul class="mt-3 space-y-2">
            {#each soon as m}
                <li class="flex items-start justify-between gap-3 text-sm">
                    <span class="min-w-0">
                        <span class="block truncate text-slate-200">{m.name}</span>
                        <span class="text-xs text-slate-500 tabular-nums">
                            Pay with {m.paikot!.cardName} {payWindow(m)} · then due {formatDate(m.paikot!.dueDate)}{Number(m.paikot!.fee) > 0 ? ` · fee ${currency.format(m.paikot!.fee)}` : ""}
                        </span>
                    </span>
                    <span class="shrink-0 text-slate-300 tabular-nums">{currency.format(m.amount)}</span>
                </li>
            {/each}
        </ul>
        {#each charges as c}
            <p class={`mt-2 text-xs tabular-nums ${c.charged > c.available ? "text-amber-300" : "text-slate-500"}`}>
                {c.card.name}: {currency.format(c.charged)} to charge · {currency.format(c.available)} available now
            </p>
        {/each}
    {:else}
        <p class="mt-3 text-sm text-slate-500">Nothing to move. Link a loan or card to the card that can pay it.</p>
    {/if}

    <details class="mt-4 border-t border-slate-800 pt-3">
        <summary class="cursor-pointer text-xs font-medium text-slate-300">Paikot links</summary>
        {#if error}
            <p role="alert" class="mt-2 rounded-lg border border-red-500/20 bg-red-500/10 px-3 py-2 text-xs text-red-300">{error}</p>
        {/if}
        {#each links as link}
            <form method="POST" action="/?/setPaikot" use:enhance class="mt-2 flex items-center justify-between gap-2 text-xs">
                <input type="hidden" name="accountId" value={link.accountId} />
                <span class="min-w-0 truncate text-slate-300">{link.accountName} → {link.paikot!.cardName}</span>
                <button class="text-slate-500 hover:text-rose-300" aria-label={`Remove the link from ${link.accountName}`}>Remove</button>
            </form>
        {/each}
        <form method="POST" action="/?/setPaikot" use:enhance class="mt-3 space-y-2">
            <label class="block text-xs text-slate-400">Dues of
                <select name="accountId" required class={field}>
                    {#each debts as a (a.id)}<option value={a.id}>{a.name}</option>{/each}
                </select>
            </label>
            <label class="block text-xs text-slate-400">Paid with
                <select name="cardId" required class={field}>
                    {#each cards as a (a.id)}<option value={a.id}>{a.name}</option>{/each}
                </select>
            </label>
            <label class="block text-xs text-slate-400">Fee (% of the amount)
                <input name="feePercent" type="number" inputmode="decimal" min="0" max="100" step="0.001" value="0" class={field} />
            </label>
            <button class="rounded-lg bg-blue-600 px-3 py-2 text-xs font-semibold text-white">Save link</button>
        </form>
    </details>
</div>

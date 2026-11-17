<script lang="ts">
    import { enhance } from "$app/forms";
    import { ChevronLeft, Plus } from "lucide-svelte";
    import AccountBadge from "$lib/components/accounts/AccountBadge.svelte";
    import AccountDeadlines from "$lib/components/accounts/AccountDeadlines.svelte";
    import InstallmentDialog from "$lib/components/accounts/InstallmentDialog.svelte";
    import LedgerTable from "$lib/components/accounts/LedgerTable.svelte";
    import { currency } from "$lib/utils/currency";
    import { formatDate, nextStatementDate } from "$lib/utils/date";
    import { cardLimits } from "$lib/utils/payables";

    let { data, form } = $props();

    let dialog: ReturnType<typeof InstallmentDialog>;
    const field = "mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white";

    const KIND: Record<string, string> = {
        ASSET: "Bank or wallet",
        LIABILITY: "Loan",
        CREDIT_CARD: "Credit card",
        RECEIVABLE: "Owed to you",
        PAYABLE: "Bill",
        INCOME: "Income category",
        EXPENSE: "Expense category"
    };

    // Loans and cards are paid down in installments.
    const owed = $derived(data.account.type === "LIABILITY" || data.account.type === "CREDIT_CARD");
    // The engine's balance is debits − credits, so money owed shows as a negative balance.
    const balance = $derived(Number(data.account.balance));
    // A card's headline is its next bill (with the installments billed on it) and when that's due, not the ledger balance.
    const nextBill = $derived.by(() => {
        const first = data.account.type === "CREDIT_CARD" ? data.deadlines[0] : undefined;
        if (!first) return null;
        const cents = data.deadlines.filter((d) => d.dueDate === first.dueDate).reduce((c, d) => c + Math.round(Number(d.amount) * 100), 0);
        return { amount: cents / 100, dueDate: first.dueDate };
    });
    // A card's limit, what's left of it (same sum as the dashboard's), and its next statement and due dates.
    const card = $derived.by(() => {
        if (data.account.type !== "CREDIT_CARD") return null;
        const limit = cardLimits([data.account], data.deadlines, "", "")[0];
        const billingDay = Number(data.account.billingDay);
        return {
            limit: limit?.limit,
            available: limit?.now,
            statement: billingDay > 0 ? nextStatementDate(billingDay, data.account.statementShift) : null,
            due: data.deadlines[0]?.dueDate
        };
    });
</script>

<svelte:head>
    <title>{data.account.name} | FinanKal</title>
</svelte:head>

<div class="space-y-10">
    <header class="space-y-6">
        <a href="/accounts"
           class="inline-flex items-center gap-1 rounded-lg text-sm text-slate-400 transition hover:text-white focus-visible:outline-2 focus-visible:outline-blue-500">
            <ChevronLeft class="h-4 w-4" /> Accounts
        </a>

        <div class="flex flex-wrap items-center justify-between gap-4">
            <div class="flex min-w-0 items-center gap-4">
                <AccountBadge name={data.account.name} type={data.account.type} />
                <div class="min-w-0">
                    <h1 class="truncate text-3xl font-bold tracking-tight text-white">{data.account.name}</h1>
                    <p class="text-sm text-slate-400">
                        {KIND[data.account.type] ?? data.account.type.toLowerCase()}, opened {formatDate(data.account.createdAt)}
                    </p>
                </div>
            </div>

            {#if owed}
                <button type="button" onclick={() => dialog.open()}
                        class="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-blue-600/20 transition hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500">
                    <Plus class="h-4 w-4" /> Add installments
                </button>
            {/if}
        </div>

        <div class="border-l-2 border-blue-500 pl-5">
            <p class="text-sm text-slate-400">{nextBill ? `You owe by ${formatDate(nextBill.dueDate)}` : owed && balance <= 0 ? "You owe" : "Balance"}</p>
            <p class={`text-4xl font-bold tracking-tight tabular-nums sm:text-5xl ${
                owed ? "text-white" : balance < 0 ? "text-red-400" : "text-white"}`}>
                {currency.format(nextBill ? nextBill.amount : owed && balance <= 0 ? -balance : balance)}
            </p>
        </div>

        {#if card}
            <dl class="grid grid-cols-2 gap-x-6 gap-y-4 rounded-2xl border border-slate-800 bg-slate-900/60 p-5 sm:grid-cols-4">
                <div>
                    <dt class="text-sm text-slate-400">Credit limit</dt>
                    <dd class="mt-1 font-semibold tabular-nums text-white">{card.limit == null ? "—" : currency.format(card.limit)}</dd>
                </div>
                <div>
                    <dt class="text-sm text-slate-400">Available credit</dt>
                    <dd class={`mt-1 font-semibold tabular-nums ${Number(card.available) < 0 ? "text-rose-400" : "text-emerald-400"}`}>{card.available == null ? "—" : currency.format(card.available)}</dd>
                </div>
                <div>
                    <dt class="text-sm text-slate-400">Next statement</dt>
                    <dd class="mt-1 font-semibold text-white">{formatDate(card.statement)}</dd>
                </div>
                <div>
                    <dt class="text-sm text-slate-400">Next due date</dt>
                    <dd class="mt-1 font-semibold text-white">{formatDate(card.due)}</dd>
                </div>
            </dl>

            <details class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5" open={!!form?.cardError}>
                <summary class="cursor-pointer text-sm font-medium text-slate-200">
                    Edit card details
                    {#if form?.cardUpdated}<span role="status" class="ml-2 text-emerald-400">Saved</span>{/if}
                </summary>
                <form method="POST" action="?/updateCard" use:enhance class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <label class="text-sm text-slate-400">Credit limit
                        <input name="creditLimit" type="number" min="0" step="0.01" required value={data.account.creditLimit ?? ""} class={field} />
                    </label>
                    <label class="text-sm text-slate-400">Statement day (1–28)
                        <input name="billingDay" type="number" min="1" max="28" required value={data.account.billingDay ?? ""} class={field} />
                    </label>
                    <label class="text-sm text-slate-400">If the statement day is a weekend
                        <select name="statementShift" value={data.account.statementShift ?? "NONE"} class={field}>
                            <option value="NONE">Keep the day</option>
                            <option value="NEXT_BUSINESS_DAY">Move to the next weekday</option>
                            <option value="NEAREST_BUSINESS_DAY">Move to the nearest weekday</option>
                        </select>
                    </label>
                    <label class="text-sm text-slate-400">Due date is
                        <select name="dueRule" value={data.account.dueRule ?? "DAYS"} class={field}>
                            <option value="DAYS">N days after the statement</option>
                            <option value="DAYS_SKIP_WEEKEND">N days after, moved off weekends</option>
                            <option value="BUSINESS_DAYS">N weekdays after the statement</option>
                            <option value="DAY_OF_MONTH">The Nth of the month</option>
                        </select>
                    </label>
                    <label class="text-sm text-slate-400">N (1–28)
                        <input name="paymentDueDays" type="number" min="1" max="28" required value={data.account.paymentDueDays ?? ""} class={field} />
                    </label>
                    <div class="flex items-end">
                        <button type="submit" class="rounded-xl bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500">Save</button>
                    </div>
                    {#if form?.cardError}<p role="alert" class="text-sm text-rose-400 sm:col-span-2 lg:col-span-3">{form.cardError}</p>{/if}
                </form>
            </details>
        {/if}
    </header>

    {#if owed || data.deadlines.length}
        <section aria-labelledby="repayment-heading" class="space-y-4">
            <div class="flex flex-wrap items-baseline justify-between gap-2">
                <h2 id="repayment-heading" class="text-lg font-semibold text-white">Repayment</h2>
                {#if form?.added}
                    <p role="status" class="text-sm text-emerald-400">
                        Added {form.added} {form.added === 1 ? "installment" : "installments"}
                    </p>
                {/if}
            </div>

            {#if data.deadlines.length}
                <AccountDeadlines deadlines={data.deadlines} isLoan={data.account.type === "LIABILITY"} isCard={data.account.type === "CREDIT_CARD"} isMayaLoan={data.installmentKind === "MAYA" && !!data.loan} payFrom={data.payFrom} linkable={data.linkablePayments} error={form?.payError ?? form?.error ?? form?.loanError} />
            {:else}
                <div class="rounded-2xl border border-dashed border-slate-700 px-6 py-10 text-center">
                    <p class="font-medium text-white">No installments yet</p>
                    <p class="mx-auto mt-1 max-w-sm text-sm text-slate-400">
                        Add the loan's payment schedule to see what's due and when it's paid off.
                    </p>
                    <button type="button" onclick={() => dialog.open()}
                            class="mt-4 inline-flex items-center gap-2 rounded-xl border border-slate-700 bg-slate-800 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-700 focus-visible:outline-2 focus-visible:outline-blue-500">
                        <Plus class="h-4 w-4" /> Add installments
                    </button>
                </div>
            {/if}
        </section>
    {/if}

    <section aria-labelledby="ledger-heading" class="space-y-4">
        <h2 id="ledger-heading" class="text-lg font-semibold text-white">Ledger</h2>
        <LedgerTable entries={data.entries} billingDay={data.account.billingDay} statementShift={data.account.statementShift} />
    </section>
</div>

{#if owed}
    <InstallmentDialog bind:this={dialog} accountName={data.account.name} payFrom={data.payFrom} defaultPayFromId={data.defaultPayFromId} defaultDisbursedAccountId={data.defaultDisbursedAccountId} feeAccounts={data.feeAccounts} isLoan={data.account.type === "LIABILITY"} disbursementAlreadyRecorded={data.disbursementAlreadyRecorded} installmentKind={data.installmentKind} statementDay={data.account.billingDay} statementShift={data.account.statementShift} error={form?.error} />
{/if}

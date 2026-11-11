<script lang="ts">
    import { ChevronLeft, Plus } from "lucide-svelte";
    import AccountBadge from "$lib/components/accounts/AccountBadge.svelte";
    import AccountDeadlines from "$lib/components/accounts/AccountDeadlines.svelte";
    import InstallmentDialog from "$lib/components/accounts/InstallmentDialog.svelte";
    import LedgerTable from "$lib/components/accounts/LedgerTable.svelte";
    import { currency } from "$lib/utils/currency";
    import { formatDate } from "$lib/utils/date";

    let { data, form } = $props();

    let dialog: ReturnType<typeof InstallmentDialog>;

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
            <p class="text-sm text-slate-400">{owed && balance <= 0 ? "You owe" : "Balance"}</p>
            <p class={`text-4xl font-bold tracking-tight tabular-nums sm:text-5xl ${
                owed ? "text-white" : balance < 0 ? "text-red-400" : "text-white"}`}>
                {currency.format(owed && balance <= 0 ? -balance : balance)}
            </p>
        </div>
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
                <AccountDeadlines deadlines={data.deadlines} isLoan={data.account.type === "LIABILITY"} isCard={data.account.type === "CREDIT_CARD"} isMayaLoan={data.installmentKind === "MAYA" && !!data.loan} payFrom={data.payFrom} error={form?.payError ?? form?.error ?? form?.loanError} />
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

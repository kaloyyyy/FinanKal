<script lang="ts">
    import TransactionForm from "$lib/components/transactions/TransactionForm.svelte";
    import TransactionEntriesTable from "$lib/components/transactions/TransactionEntriesTable.svelte";
    import TransactionSummary from "$lib/components/transactions/TransactionSummary.svelte";
    import TransactionActions from "$lib/components/transactions/TransactionActions.svelte";

    let {data, form} = $props();
    let description = $state("");
    let transactionDate = $state(
        new Date().toISOString().split("T")[0]
    );

    let loading = $state(false);

    let entries = $state([
        {
            accountId: "",
            amount: 0,
            type: "DEBIT"
        },
        {
            accountId: "",
            amount: 0,
            type: "CREDIT"
        }
    ]);

    const totalDebit = $derived(
        entries
            .filter((e) => e.type === "DEBIT")
            .reduce((s, e) => s + Number(e.amount || 0), 0)
    );

    const totalCredit = $derived(
        entries
            .filter((e) => e.type === "CREDIT")
            .reduce((s, e) => s + Number(e.amount || 0), 0)
    );

    const canSubmit = $derived(
        description.trim().length > 0 &&
        entries.length >= 2 &&
        totalDebit === totalCredit &&
        totalDebit > 0 &&
        entries.every(
            (e) =>
                e.accountId !== "" &&
                e.amount > 0
        )
    );

    async function save() {
        loading = true;

        try {
            // await createTransaction(...)
        } finally {
            loading = false;
        }
    }
</script>

<div class="mx-auto max-w-6xl space-y-6">

    <form method="POST" action="?/create">

        <TransactionForm
                bind:description
                bind:transactionDate
        />

        <TransactionEntriesTable
                bind:entries
                accounts={data.accounts}
        />

        <TransactionSummary
                {entries}
        />

        <input
                type="hidden"
                name="transaction"
                value={JSON.stringify({
            description,
            entries
        })}
        />

        <TransactionActions
                {canSubmit}
                loading={false}
                onCancel={() => history.back()}
        />

    </form>
    {#if form?.success}
        <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4">
            <div class="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
                <div class="flex items-center gap-4">
                    <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-emerald-500/10 text-2xl">
                        ✓
                    </div>

                    <div>
                        <h2 class="text-xl font-semibold text-white">
                            Transaction Created
                        </h2>

                        <p class="mt-1 text-sm text-slate-400">
                            Your transaction has been successfully recorded.
                        </p>
                    </div>

                </div>

                <div class="mt-6 flex justify-end gap-3">

                    {#if form?.success}
                        <p class="mt-4 text-xs text-slate-500">
                            Transaction ID:
                            <span class="font-mono text-slate-400">
                                {form.transaction?.id}
                            </span>
                        </p>
                    {/if}

                    <a href="/transactions/new" class="rounded-xl bg-blue-600 px-5 py-2.5 font-semibold text-white transition hover:bg-blue-500">
                        New Transaction
                    </a>

                </div>
            </div>
        </div>
    {/if}
</div>
<script lang="ts">
    // Replace this with your server-loaded data once the GET API is connected.
    let { data } = $props();

    const currency = new Intl.NumberFormat("en-PH", {
        style: "currency",
        currency: "PHP"
    });

    function formatDate(date: string | null | undefined) {
        if (!date) return "—";

        return new Date(date).toLocaleDateString("en-PH", {
            year: "numeric",
            month: "short",
            day: "numeric"
        });
    }

    function formatType(type: string) {
        return type === "DEBIT" ? "Debit" : "Credit";
    }

    function typeClass(type: string) {
        return type === "DEBIT"
            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
            : "bg-blue-500/10 text-blue-400 border-blue-500/20";
    }
</script>

<svelte:head>
    <title>Transactions | FinanKal</title>
</svelte:head>

<div class="space-y-6">

    <!-- Header -->

    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">

        <div>
            <h1 class="text-3xl font-bold tracking-tight text-white">
                Transactions
            </h1>

            <p class="mt-1 text-slate-400">
                View and manage your financial transactions.
            </p>
        </div>

        <a
                href="/transactions/new"
                class="inline-flex items-center justify-center rounded-xl bg-blue-600 px-5 py-3 font-semibold text-white transition hover:bg-blue-500"
        >
            + New Transaction
        </a>

    </div>


    <!-- Summary -->

    <div class="grid gap-4 md:grid-cols-3">

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Transactions
            </p>

            <p class="mt-2 text-3xl font-bold text-white">
                {data.transactions?.length ?? 0}
            </p>
        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Recent Activity
            </p>

            <p class="mt-2 text-3xl font-bold text-emerald-400">
                Active
            </p>
        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Ledger
            </p>

            <p class="mt-2 text-3xl font-bold text-blue-400">
                Double Entry
            </p>
        </div>

    </div>


    <!-- Transactions -->

    <div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

        <div class="border-b border-slate-800 px-5 py-4">
            <h2 class="font-semibold text-white">
                All Transactions
            </h2>
        </div>

        {#if !data.transactions || data.transactions.length === 0}

            <div class="px-6 py-16 text-center">

                <div class="mx-auto mb-4 text-5xl">
                    💸
                </div>

                <h3 class="text-lg font-semibold text-white">
                    No transactions yet
                </h3>

                <p class="mt-2 text-sm text-slate-400">
                    Create your first transaction to start tracking your finances.
                </p>

                <a
                        href="/transactions/new"
                        class="mt-6 inline-flex rounded-xl bg-blue-600 px-5 py-3 font-medium text-white transition hover:bg-blue-500"
                >
                    Create Transaction
                </a>

            </div>

        {:else}

            <div class="overflow-x-auto">

                <table class="w-full text-left">

                    <thead class="border-b border-slate-800 bg-slate-950/50">

                    <tr>
                        <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                            Date
                        </th>

                        <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                            Description
                        </th>

                        <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                            Entries
                        </th>

                        <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-500">
                            Amount
                        </th>

                        <th class="px-5 py-4"></th>
                    </tr>

                    </thead>

                    <tbody class="divide-y divide-slate-800">

                    {#each data.transactions as transaction}

                        <tr class="transition hover:bg-slate-800/40">

                            <td class="whitespace-nowrap px-5 py-4 text-sm text-slate-400">
                                {formatDate(transaction.createdAt)}
                            </td>

                            <td class="px-5 py-4">

                                <a
                                        href={`/transactions/${transaction.id}`}
                                        class="font-medium text-white transition hover:text-blue-400"
                                >
                                    {transaction.description}
                                </a>

                            </td>

                            <td class="px-5 py-4">

                                    <span
                                            class="inline-flex rounded-full border border-slate-700 bg-slate-800 px-2.5 py-1 text-xs font-medium text-slate-300"
                                    >
                                        {transaction.entries?.length ?? 0}
                                        {transaction.entries?.length === 1 ? " entry" : " entries"}
                                    </span>

                            </td>

                            <td class="px-5 py-4 text-right font-semibold text-white">

                                {#if transaction.entries?.length}
                                    {currency.format(transaction.entries[0].amount)}
                                {:else}
                                    —
                                {/if}

                            </td>

                            <td class="px-5 py-4 text-right">

                                <a
                                        href={`/transactions/${transaction.id}`}
                                        class="text-sm font-medium text-blue-400 transition hover:text-blue-300"
                                >
                                    View →
                                </a>

                            </td>

                        </tr>

                    {/each}

                    </tbody>

                </table>

            </div>

        {/if}

    </div>

</div>
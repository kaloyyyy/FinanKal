<script lang="ts">
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

    function formatTime(date: string | null | undefined) {
        if (!date) return "";

        return new Date(date).toLocaleTimeString("en-PH", {
            hour: "numeric",
            minute: "2-digit"
        });
    }

    function getTransactionAmount(transaction: any): number {
        if (!transaction.entries?.length) {
            return 0;
        }

        // A double-entry transaction has equal debit/credit totals.
        // Use the debit side as the displayed transaction amount.
        const debit = transaction.entries
            .filter((entry: any) => entry.type === "DEBIT")
            .reduce(
                (total: number, entry: any) =>
                    total + Number(entry.amount),
                0
            );

        return debit;
    }

    function getEntrySummary(transaction: any): string {
        if (!transaction.entries?.length) {
            return "No entries";
        }

        return transaction.entries
            .map((entry: any) => entry.accountId)
            .join(" → ");
    }

    function getEntryTypes(transaction: any): string {
        if (!transaction.entries?.length) {
            return "—";
        }

        const debits = transaction.entries.filter(
            (entry: any) => entry.type === "DEBIT"
        ).length;

        const credits = transaction.entries.filter(
            (entry: any) => entry.type === "CREDIT"
        ).length;

        return `${debits} debit${debits !== 1 ? "s" : ""} · ${credits} credit${credits !== 1 ? "s" : ""}`;
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
                {data.total ?? data.transactions?.length ?? 0}
            </p>

            <p class="mt-1 text-xs text-slate-500">
                Recorded in your ledger
            </p>
        </div>


        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Current Page
            </p>

            <p class="mt-2 text-3xl font-bold text-blue-400">
                {data.transactions?.length ?? 0}
            </p>

            <p class="mt-1 text-xs text-slate-500">
                Transactions displayed
            </p>
        </div>


        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Ledger
            </p>

            <p class="mt-2 text-3xl font-bold text-emerald-400">
                Double Entry
            </p>

            <p class="mt-1 text-xs text-slate-500">
                Every transaction is balanced
            </p>
        </div>

    </div>


    <!-- Transactions -->

    <div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

        <div class="flex flex-col gap-3 border-b border-slate-800 px-5 py-4 sm:flex-row sm:items-center sm:justify-between">

            <div>
                <h2 class="font-semibold text-white">
                    All Transactions
                </h2>

                <p class="mt-1 text-xs text-slate-500">
                    Your double-entry ledger activity
                </p>
            </div>

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

                        {@const amount = getTransactionAmount(transaction)}

                        <tr class="transition hover:bg-slate-800/40">

                            <!-- Date -->

                            <td class="whitespace-nowrap px-5 py-4">

                                <div class="text-sm font-medium text-slate-300">
                                    {formatDate(transaction.createdAt)}
                                </div>

                                <div class="mt-0.5 text-xs text-slate-500">
                                    {formatTime(transaction.createdAt)}
                                </div>

                            </td>


                            <!-- Description -->

                            <td class="px-5 py-4">

                                <a
                                        href={`/transactions/${transaction.id}`}
                                        class="font-medium text-white transition hover:text-blue-400"
                                >
                                    {transaction.description || "Untitled transaction"}
                                </a>

                                <p class="mt-1 max-w-md truncate text-xs text-slate-500">
                                    {getEntrySummary(transaction)}
                                </p>

                            </td>


                            <!-- Entries -->

                            <td class="px-5 py-4">

                                <div class="flex flex-col gap-1">

                                    <span
                                            class="inline-flex w-fit rounded-full border border-slate-700 bg-slate-800 px-2.5 py-1 text-xs font-medium text-slate-300"
                                    >
                                        {transaction.entries?.length ?? 0}
                                        {transaction.entries?.length === 1
                                            ? " entry"
                                            : " entries"}
                                    </span>

                                    <span class="text-xs text-slate-500">
                                        {getEntryTypes(transaction)}
                                    </span>

                                </div>

                            </td>


                            <!-- Amount -->

                            <td class="px-5 py-4 text-right">

                                {#if transaction.entries?.length}

                                    <span class="font-semibold text-white">
                                        {currency.format(amount)}
                                    </span>

                                {:else}

                                    <span class="text-slate-500">
                                        —
                                    </span>

                                {/if}

                            </td>


                            <!-- Action -->

                            <td class="px-5 py-4 text-right">

                                <a
                                        href={`/transactions/${transaction.id}`}
                                        class="inline-flex items-center rounded-lg px-3 py-2 text-sm font-medium text-blue-400 transition hover:bg-blue-500/10 hover:text-blue-300"
                                >
                                    View
                                    <span class="ml-1">
                                        →
                                    </span>
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
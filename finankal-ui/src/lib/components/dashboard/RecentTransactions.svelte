<script lang="ts">
    import type { RecentTransaction } from "$lib/server/api/transactions";
    import {currency} from "$lib/utils/currency.ts";
    let {
        transactions = []
    }: {
        transactions: RecentTransaction[];
    } = $props();

    function formatDate(date: string) {
        const value = new Date(date);

        return value.toLocaleDateString("en-PH", {
            month: "short",
            day: "numeric"
        });
    }

    function amountClass(direction: RecentTransaction["direction"]) {
        if (direction === "IN") {
            return "text-emerald-400";
        }

        if (direction === "OUT") {
            return "text-red-400";
        }

        return "text-slate-300";
    }

    function amountPrefix(
        direction: RecentTransaction["direction"]
    ) {
        if (direction === "IN") {
            return "+";
        }

        if (direction === "OUT") {
            return "-";
        }

        return "";
    }
</script>

<div class="rounded-2xl border border-slate-800 bg-slate-900">

    <div
            class="flex items-center justify-between border-b border-slate-800 px-5 py-4"
    >

        <div>
            <h2 class="text-lg font-semibold text-white">
                Recent Transactions
            </h2>

            <p class="mt-1 text-sm text-slate-400">
                Your latest financial activity
            </p>
        </div>

        <a
                href="/transactions"
                class="text-sm font-medium text-blue-400 transition hover:text-blue-300"
        >
            View All →
        </a>

    </div>

    {#if transactions.length === 0}

        <div class="px-5 py-12 text-center">

            <div class="text-4xl">
                🧾
            </div>

            <p class="mt-3 font-medium text-white">
                No transactions yet
            </p>

            <p class="mt-1 text-sm text-slate-500">
                Create your first transaction to see it here.
            </p>

            <a
                    href="/transactions/new"
                    class="mt-5 inline-flex rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-500"
            >
                + New Transaction
            </a>

        </div>

    {:else}

        <div class="divide-y divide-slate-800">

            {#each transactions as transaction}

                <a
                        href={`/transactions/${transaction.id}`}
                        class="flex items-center justify-between px-5 py-4 transition hover:bg-slate-800/50"
                >

                    <div class="flex min-w-0 items-center gap-3">

                        <div
                                class={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${
                                transaction.direction === "IN"
                                    ? "bg-emerald-500/10"
                                    : transaction.direction === "OUT"
                                        ? "bg-red-500/10"
                                        : "bg-slate-800"
                            }`}
                        >
                            {#if transaction.direction === "IN"}
                                ↓
                            {:else if transaction.direction === "OUT"}
                                ↑
                            {:else}
                                ↔
                            {/if}
                        </div>

                        <div class="min-w-0">

                            <p class="truncate font-medium text-white">
                                {transaction.description}
                            </p>

                            <div class="mt-1 flex items-center gap-2 text-xs text-slate-500">

                                <span>
                                    {transaction.accountName}
                                </span>

                                <span>•</span>

                                <span>
                                    {formatDate(transaction.createdAt)}
                                </span>

                            </div>

                        </div>

                    </div>

                    <div
                            class={`ml-4 shrink-0 text-right font-semibold ${amountClass(transaction.direction)}`}
                    >
                        {amountPrefix(transaction.direction)}
                        {currency.format(transaction.amount)}
                    </div>

                </a>

            {/each}

        </div>

    {/if}

</div>
<script lang="ts">
    import { formatDate } from "$lib/utils/date";
    import type { RecentTransaction } from "$lib/server/api/transactions";
    import {currency} from "$lib/utils/currency.ts";
    import { ArrowDownLeft, ArrowUpRight, ArrowLeftRight, Receipt, Plus } from "lucide-svelte";
    let {
        transactions = []
    }: {
        transactions: RecentTransaction[];
    } = $props();

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

            <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-slate-800 text-slate-400">
                <Receipt class="h-6 w-6" />
            </div>

            <p class="mt-3 font-medium text-white">
                No transactions yet
            </p>

            <p class="mt-1 text-sm text-slate-500">
                Create your first transaction to see it here.
            </p>

            <a
                    href="/transactions/new"
                    class="mt-5 inline-flex items-center gap-2 rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-500"
            >
                <Plus class="h-4 w-4" />
                New Transaction
            </a>

        </div>

    {:else}

        <div class="hidden grid-cols-[2.5rem_1fr_160px_100px_140px] gap-4 border-b border-slate-800 px-5 py-2 text-xs font-medium uppercase tracking-wide text-slate-500 md:grid">
            <span></span>
            <span>Description</span>
            <span>Account</span>
            <span>Date</span>
            <span class="text-right">Amount</span>
        </div>

        <div class="divide-y divide-slate-800">

            {#each transactions as transaction}

                <a
                        href={`/transactions/${transaction.id}`}
                        class="grid grid-cols-[2.5rem_1fr_auto] items-center gap-3 px-5 py-4 transition hover:bg-slate-800/50 md:grid-cols-[2.5rem_1fr_160px_100px_140px] md:gap-4"
                >

                    <div
                            class={`flex h-10 w-10 items-center justify-center rounded-full ${
                            transaction.direction === "IN"
                                ? "bg-emerald-500/10 text-emerald-400"
                                : transaction.direction === "OUT"
                                    ? "bg-red-500/10 text-red-400"
                                    : "bg-slate-800 text-slate-300"
                        }`}
                    >
                        {#if transaction.direction === "IN"}
                            <ArrowDownLeft class="h-5 w-5" />
                        {:else if transaction.direction === "OUT"}
                            <ArrowUpRight class="h-5 w-5" />
                        {:else}
                            <ArrowLeftRight class="h-5 w-5" />
                        {/if}
                    </div>

                    <div class="min-w-0">
                        <p class="truncate font-medium text-white">
                            {transaction.description}
                        </p>

                        <!-- account/date inline on mobile; own columns on md+ -->
                        <p class="mt-1 truncate text-xs text-slate-500 md:hidden">
                            {transaction.accountName} • {formatDate(transaction.createdAt)}
                        </p>
                    </div>

                    <span class="hidden truncate text-sm text-slate-400 md:block">
                        {transaction.accountName}
                    </span>

                    <span class="hidden text-sm text-slate-500 md:block">
                        {formatDate(transaction.createdAt)}
                    </span>

                    <div class={`text-right font-semibold tabular-nums ${amountClass(transaction.direction)}`}>
                        {amountPrefix(transaction.direction)}{currency.format(transaction.amount)}
                    </div>

                </a>

            {/each}

        </div>

    {/if}

</div>
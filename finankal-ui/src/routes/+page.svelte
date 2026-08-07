<script lang="ts">
    import SummaryCards from "$lib/components/dashboard/SummaryCards.svelte";

    let { data } = $props();
</script>


<div class="mx-auto flex w-full max-w-7xl flex-col gap-4">

    <SummaryCards
            totalCredit={data.summary.totalCredit}
            totalDebit={data.summary.totalDebit}
            netWorth={data.summary.netWorth}
    />

    <div class="rounded-2xl border border-slate-800 bg-slate-900 p-6">
        <div class="flex flex-col gap-5 md:flex-row md:items-center md:justify-between">

            <div>
                <h2 class="text-xl font-semibold text-white">
                    Your Accounts
                </h2>

                <p class="mt-1 max-w-2xl text-sm text-slate-400">
                    View and manage your bank accounts, credit cards, loans, and digital wallets.
                </p>
            </div>

            <div class="flex shrink-0 flex-col gap-3 sm:flex-row">

                <a
                        href="/transactions/new"
                        class="inline-flex items-center justify-center rounded-xl bg-blue-600 px-5 py-3 font-semibold text-white transition hover:bg-blue-500"
                >
                    <span class="mr-2 text-lg">+</span>
                    New Transaction
                </a>

                <a
                        href="/accounts"
                        class="inline-flex items-center justify-center rounded-xl border border-slate-700 bg-slate-800 px-5 py-3 font-medium text-slate-300 transition hover:border-slate-600 hover:bg-slate-700 hover:text-white"
                >
                    View Accounts →
                </a>

            </div>

        </div>
    </div>

</div>
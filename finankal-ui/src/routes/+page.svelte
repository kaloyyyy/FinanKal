<script lang="ts">
    import SummaryCards from "$lib/components/dashboard/SummaryCards.svelte";
    import RecentTransactions from "$lib/components/dashboard/RecentTransactions.svelte";
    import AccountsPanel from "$lib/components/dashboard/AccountsPanel.svelte";
    import NetWorthChart from "$lib/components/dashboard/NetWorthChart.svelte";
    import PayablesCard from "$lib/components/dashboard/PayablesCard.svelte";

    let { data } = $props();
</script>

<svelte:head>
    <title>Dashboard | FinanKal</title>
</svelte:head>

<div class="flex flex-col gap-6">

    <div>
        <h1 class="text-[28px] font-bold tracking-tight text-white">
            Dashboard
        </h1>

        <p class="mt-1 text-sm text-slate-400">
            At a glance: what you have, what you owe, and your net worth.
        </p>
    </div>

    <SummaryCards
            totalCredit={data.summary.totalCredit}
            totalDebit={data.summary.totalDebit}
            netWorth={data.summary.netWorth}
    />

    {#if data.netPosition}
        <PayablesCard {...data.netPosition} />
    {/if}

    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div class="lg:col-span-2">
            <NetWorthChart points={data.netWorthHistory} />
        </div>

        <AccountsPanel accounts={data.accounts} />
    </div>

    <RecentTransactions
            transactions={data.recentTransactions}
    />

</div>

<script lang="ts">
    import AccountCard from "$lib/components/accounts/AccountCard.svelte";
    import AccountBalance from "$lib/components/accounts/AccountBalance.svelte";
    import LedgerTable from "$lib/components/accounts/LedgerTable.svelte";

    let { data } = $props();
</script>

<svelte:head>
    <title>{data.account.name} | FinanKal</title>
</svelte:head>

<div class="space-y-8">

    <div>
        <h1 class="text-3xl font-bold">
            {data.account.name}
        </h1>

        <p class="text-base-content/70">
            {data.account.type}
        </p>
    </div>

    <AccountBalance balance={data.account.balance} />

    <AccountCard account={data.account} />

    <div class="card bg-base-100 shadow">

        <div class="card-body">

            <h2 class="card-title">
                Ledger Entries
            </h2>

            <LedgerTable entries={data.entries} />

        </div>

    </div>

</div>
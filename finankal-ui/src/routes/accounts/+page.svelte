<script lang="ts">
    import AccountCard from '$lib/components/accounts/AccountCard.svelte';

    let { data } = $props();

    const grouped = Object.groupBy(
        data.accounts,
        (account: any) => account.type
    );
</script>

<svelte:head>
    <title>Accounts | FinanKal</title>
</svelte:head>

<div class="space-y-8 mx-auto flex w-full max-w-7xl flex-col gap-4">

    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
            <h1 class="text-3xl font-bold tracking-tight text-white">
                Accounts
            </h1>

            <p class="mt-1 text-slate-400">
                Manage your bank accounts, credit cards, loans, and digital wallets.
            </p>
        </div>

        <button
                class="rounded-xl bg-blue-600 px-5 py-3 font-medium text-white transition hover:bg-blue-500"
        >
            + New Account
        </button>
    </div>

    <div class="grid gap-4 md:grid-cols-3">

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Total Accounts
            </p>

            <p class="mt-2 text-3xl font-bold text-white">
                {data.accounts.length}
            </p>
        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Asset Accounts
            </p>

            <p class="mt-2 text-3xl font-bold text-emerald-400">
                {grouped.ASSET?.length ?? 0}
            </p>
        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">
            <p class="text-sm text-slate-400">
                Liability Accounts
            </p>

            <p class="mt-2 text-3xl font-bold text-red-400">
                {(grouped.CREDIT_CARD?.length ?? 0) + (grouped.LIABILITY?.length ?? 0)}
            </p>
        </div>

    </div>

    {#if data.accounts.length === 0}

        <div class="rounded-2xl border border-dashed border-slate-700 bg-slate-900 py-20 text-center">

            <div class="mx-auto mb-4 text-6xl">
                🏦
            </div>

            <h2 class="text-xl font-semibold text-white">
                No accounts yet
            </h2>

            <p class="mt-2 text-slate-400">
                Create your first account to start tracking your finances.
            </p>

            <button
                    class="mt-6 rounded-xl bg-blue-600 px-6 py-3 font-medium text-white transition hover:bg-blue-500"
            >
                Create Account
            </button>

        </div>

    {:else}

        {#each Object.entries(grouped).filter(([type]) => type !== "CLEARING" && type !== "EQUITY") as [type, accounts]}

            <section class="space-y-4">

                <div class="flex items-center space-x-4">

                    <h2 class="text-xl font-semibold text-white">
                        {type.replaceAll("_", " ")}
                    </h2>

                    <span
                            class="rounded-full border border-slate-700 bg-slate-800 px-3 py-1 text-sm text-slate-300"
                    >
						{accounts.length}
					</span>

                </div>

                <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                    {#each accounts as account}
                        <AccountCard {account}/>
                    {/each}
                </div>

            </section>

        {/each}

    {/if}

</div>
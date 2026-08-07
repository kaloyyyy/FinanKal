<script lang="ts">
    import AccountTypeBadge from './AccountTypeBadge.svelte';

    let { account } = $props();

    const currency = new Intl.NumberFormat('en-PH', {
        style: 'currency',
        currency: 'PHP'
    });

    const balanceClass =
        account.balance < 0
            ? 'text-red-600'
            : 'text-green-600';
</script>
{#if account.type !== "CLEARING" && account.type !== "EQUITY"}
    <a
            href={`/accounts/${account.id}`}
            class="block rounded-2xl border border-slate-800 bg-slate-900 p-5 transition hover:border-blue-500 hover:bg-slate-800 hover:shadow-xl"
    >
        <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
                <div
                        class="flex h-12 w-12 items-center justify-center rounded-full bg-slate-800 text-xl"
                >
                    {#if account.type === "ASSET"}
                        💰
                    {:else if account.type === "CREDIT_CARD"}
                        💳
                    {:else if account.type === "LIABILITY"}
                        📄
                    {:else}
                        📁
                    {/if}
                </div>

                <div>
                    <h3 class="font-semibold text-white">
                        {account.name}
                    </h3>

                    <div class="mt-1">
                        <AccountTypeBadge type={account.type} />
                    </div>
                </div>
            </div>

            <div class="text-right">
                <div class={`text-xl font-bold ${balanceClass}`}>
                    {currency.format(account.balance)}
                </div>

                <p class="mt-1 text-xs text-slate-400">
                    Current Balance
                </p>
            </div>
        </div>

        <div class="mt-6 flex items-center justify-between border-t border-slate-800 pt-4 text-sm text-slate-400">
			<span>
				Created {new Date(account.createdAt).toLocaleDateString()}
			</span>

            <span class="font-medium text-blue-400">
				View →
			</span>
        </div>
    </a>
{/if}
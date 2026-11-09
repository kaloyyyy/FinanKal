<script lang="ts">
    import type { Account } from "$lib/server/api/accounts";
    import AccountTypeBadge from "$lib/components/accounts/AccountTypeBadge.svelte";
    import { currency } from "$lib/utils/currency.ts";
    import { Landmark, CreditCard, Receipt, Folder, ChevronRight } from "lucide-svelte";

    let { accounts = [] }: { accounts: Account[] } = $props();

    // Internal ledger accounts are not user-facing.
    const visible = $derived(
        accounts.filter((a) => a.type !== "CLEARING" && a.type !== "EQUITY").slice(0, 5)
    );

    const icons = { ASSET: Landmark, CREDIT_CARD: CreditCard, LIABILITY: Receipt } as Record<string, typeof Landmark>;
</script>

<div class="flex flex-col rounded-2xl border border-slate-800 bg-slate-900 p-6">
    <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-white">Accounts</h2>
        <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-slate-800 text-blue-400">
            <Landmark class="h-4 w-4" />
        </div>
    </div>

    {#if visible.length === 0}
        <p class="flex-1 py-10 text-center text-sm text-slate-500">No accounts yet.</p>
    {:else}
        <div class="mt-4 flex flex-1 flex-col divide-y divide-slate-800">
            {#each visible as account (account.id)}
                {@const Icon = icons[account.type] ?? Folder}
                <a
                        href={`/accounts/${account.id}`}
                        class="-mx-2 flex items-center justify-between gap-3 rounded-lg px-2 py-3 transition hover:bg-slate-800/50"
                >
                    <div class="flex min-w-0 items-center gap-3">
                        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-800 text-slate-300">
                            <Icon class="h-4 w-4" />
                        </div>
                        <div class="min-w-0">
                            <p class="truncate text-sm font-medium text-white">{account.name}</p>
                            <div class="mt-0.5"><AccountTypeBadge type={account.type} /></div>
                        </div>
                    </div>
                    <div class={`shrink-0 text-right text-sm font-semibold tabular-nums ${account.balance < 0 ? "text-red-400" : "text-white"}`}>
                        {currency.format(account.balance)}
                    </div>
                </a>
            {/each}
        </div>
    {/if}

    <a
            href="/accounts"
            class="mt-4 flex items-center justify-center gap-1 rounded-xl border border-slate-700 bg-slate-800 px-4 py-2.5 text-sm font-medium text-slate-200 transition hover:bg-slate-700"
    >
        View all accounts
        <ChevronRight class="h-4 w-4" />
    </a>
</div>

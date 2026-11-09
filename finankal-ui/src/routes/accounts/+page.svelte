<script lang="ts">
    import type { Account } from '$lib/server/api/accounts';
    import { currency } from '$lib/utils/currency.ts';
    import { Landmark, ChevronRight } from 'lucide-svelte';
    import AccountBadge from '$lib/components/accounts/AccountBadge.svelte';
    import NetPositionCarousel from '$lib/components/accounts/NetPositionCarousel.svelte';
    import { formatMonth } from '$lib/utils/date';
    import { positionTiles, slicesByAccount } from '$lib/utils/payables';

    let { data } = $props();


    // Balances come from the engine as debit − credit, so what you owe is negative.
    const OWED = ['CREDIT_CARD', 'LIABILITY', 'PAYABLE'];
    const HELD = ['ASSET', 'RECEIVABLE'];
    const HIDDEN = ['CLEARING', 'EQUITY'];

    const typeName = (t: string) => t.charAt(0) + t.slice(1).toLowerCase().replaceAll('_', ' ');

    const accounts = $derived(
        (data.accounts as Account[])
            .filter((a) => !HIDDEN.includes(a.type))
            .map((a) => ({ ...a, balance: Number(a.balance) }))
    );

    const held = $derived(accounts.filter((a) => HELD.includes(a.type)).map((a) => ({ ...a, amount: a.balance })));
    const owed = $derived(accounts.filter((a) => OWED.includes(a.type)).map((a) => ({ ...a, amount: -a.balance })));
    const other = $derived(accounts.filter((a) => !HELD.includes(a.type) && !OWED.includes(a.type)).map((a) => ({ ...a, amount: a.balance })));

    const sum = (list: { amount: number }[]) => list.reduce((t, a) => t + a.amount, 0);
    const totalHeld = $derived(sum(held));
    const totalOwed = $derived(sum(owed));
    const net = $derived(totalHeld - totalOwed);

    // Only positive amounts take up room in the split bar.
    const barTotal = $derived(sum(held.filter((a) => a.amount > 0)) + sum(owed.filter((a) => a.amount > 0)));
    const width = (amount: number) => `${(Math.max(amount, 0) / barTotal) * 100}%`;

    // Where you stand: if you paid every debt today, after this month's bills, after next month's.
    // Null if payables couldn't be loaded: the single balance panel shows instead.
    const positionSlides = $derived.by(() => {
        if (!data.netPosition) return null;
        const { positions, thisMonth, nextMonth, salaryInto } = data.netPosition;
        const tiles = positionTiles(positions, thisMonth.month, nextMonth.month, (m) => formatMonth(`${m}-01`));
        const byAccount = held.filter((a) => a.type === 'ASSET');
        // Next month's salary, shown under the account it's paid into (UnionBank Payroll today).
        const salary = salaryInto && positions.salary > 0
            ? [{ id: `salary-${salaryInto.id}`, name: salaryInto.name, label: `Predicted salary (${salaryInto.name})`, amount: positions.salary, predicted: true }]
            : undefined;
        // Every slide starts from your bank and wallet balances; what's owed or due is split by the account it pays down.
        const slices = [
            { 0: byAccount, 1: owed },
            { 0: byAccount, 1: slicesByAccount(thisMonth.items) },
            { 0: byAccount, ...(salary ? { 1: salary } : {}), 2: slicesByAccount(nextMonth.items) }
        ];
        return tiles.map((tile, i) => ({ tile, slices: slices[i] }));
    });

    const groups = $derived([
        { title: 'What you have', items: held, total: totalHeld },
        { title: 'What you owe', items: owed, total: totalOwed },
        { title: 'Other accounts', items: other, total: null }
    ].filter((g) => g.items.length > 0));
</script>

<svelte:head>
    <title>Accounts | FinanKal</title>
</svelte:head>

<div class="space-y-10">
    <header>
        <h1 class="text-3xl font-bold tracking-tight text-white">Accounts</h1>
        <p class="mt-1 text-slate-400">Every account you track, and where you stand once debts are paid off.</p>
    </header>

    {#if accounts.length === 0}
        <div class="rounded-2xl border border-dashed border-slate-700 px-6 py-16 text-center">
            <Landmark class="mx-auto h-8 w-8 text-slate-500" />
            <h2 class="mt-4 text-lg font-semibold text-white">No accounts yet</h2>
            <p class="mx-auto mt-2 max-w-sm text-sm text-slate-400">
                Bank accounts, wallets, credit cards and loans show up here once they're added.
            </p>
        </div>
    {:else}
        {#if positionSlides}
            <NetPositionCarousel slides={positionSlides} />
        {:else}
        <!-- Balance sheet: have − owe = net, with each account as a slice of the bar -->
        <section aria-labelledby="position" class="rounded-2xl border border-slate-800 bg-slate-900/60 p-6 sm:p-8">
            <h2 id="position" class="sr-only">Your position</h2>

            <dl class="grid grid-cols-[1fr_auto_1fr] items-end gap-x-3 gap-y-6 sm:grid-cols-[1fr_auto_1fr_auto_1.4fr] sm:gap-x-6">
                <div>
                    <dt class="text-sm text-slate-400">You have</dt>
                    <dd class="mt-1 text-xl font-semibold tabular-nums text-emerald-400 sm:text-2xl">{currency.format(totalHeld)}</dd>
                </div>
                <span aria-hidden="true" class="pb-1 text-2xl font-light text-slate-600">−</span>
                <div>
                    <dt class="text-sm text-slate-400">You owe</dt>
                    <dd class="mt-1 text-xl font-semibold tabular-nums text-rose-400 sm:text-2xl">{currency.format(totalOwed)}</dd>
                </div>
                <span aria-hidden="true" class="hidden pb-2 text-3xl font-light text-slate-600 sm:block">=</span>
                <div class="col-span-3 border-t border-slate-800 pt-5 sm:col-span-1 sm:border-0 sm:pt-0">
                    <dt class="text-sm text-slate-400">Net position</dt>
                    <dd class={`mt-1 text-4xl font-bold tracking-tight tabular-nums sm:text-5xl ${net < 0 ? 'text-rose-300' : 'text-white'}`}>
                        {currency.format(net)}
                    </dd>
                </div>
            </dl>

            {#if barTotal > 0}
                <div class="mt-8 flex h-3 gap-0.5 overflow-hidden rounded-full bg-slate-800" role="img"
                     aria-label={`${currency.format(totalHeld)} held against ${currency.format(totalOwed)} owed`}>
                    {#each held.filter((a) => a.amount > 0) as a (a.id)}
                        <span class="bg-emerald-500 odd:bg-emerald-400" style:width={width(a.amount)} title={`${a.name}: ${currency.format(a.amount)}`}></span>
                    {/each}
                    {#each owed.filter((a) => a.amount > 0) as a (a.id)}
                        <span class="bg-rose-500 odd:bg-rose-400" style:width={width(a.amount)} title={`${a.name}: ${currency.format(a.amount)}`}></span>
                    {/each}
                </div>
                <p class="mt-3 text-sm text-slate-500">
                    {#if totalOwed <= 0}
                        Nothing owed right now.
                    {:else if totalHeld <= 0}
                        Debts aren't covered by anything you hold yet.
                    {:else}
                        Your debts equal {Math.round((totalOwed / totalHeld) * 100)}% of what you hold.
                    {/if}
                </p>
            {/if}
        </section>
        {/if}

        {#each groups as group, i (group.title)}
            <section aria-labelledby={`group-${i}`}>
                <div class="flex items-baseline justify-between border-b border-slate-800 pb-3">
                    <h2 id={`group-${i}`} class="text-lg font-semibold text-white">
                        {group.title}
                        <span class="ml-1 text-sm font-normal text-slate-500">{group.items.length}</span>
                    </h2>
                    {#if group.total !== null}
                        <span class="text-sm font-medium tabular-nums text-slate-300">{currency.format(group.total)}</span>
                    {/if}
                </div>

                <ul class="divide-y divide-slate-800/70">
                    {#each group.items as account (account.id)}
                        <li>
                            <a href={`/accounts/${account.id}`}
                               class="group -mx-3 flex items-center gap-4 rounded-xl px-3 py-4 transition-colors hover:bg-slate-900 focus-visible:bg-slate-900 focus-visible:outline-2 focus-visible:outline-blue-500">
                                <AccountBadge name={account.name} type={account.type} />
                                <span class="min-w-0 flex-1">
                                    <span class="block truncate font-medium text-white">{account.name}</span>
                                    <span class="block text-sm text-slate-500">
                                        {typeName(account.type)}, since {new Date(account.createdAt).toLocaleDateString('en-PH', { month: 'short', year: 'numeric' })}
                                    </span>
                                </span>
                                <span class={`shrink-0 text-right font-semibold tabular-nums ${account.amount < 0 ? 'text-rose-400' : 'text-white'}`}>
                                    {currency.format(account.amount)}
                                </span>
                                <ChevronRight class="h-4 w-4 shrink-0 text-slate-600 transition-colors group-hover:text-slate-300" />
                            </a>
                        </li>
                    {/each}
                </ul>
            </section>
        {/each}
    {/if}
</div>

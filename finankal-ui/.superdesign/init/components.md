# Components (finankal-ui)
Stack: SvelteKit 2 + Svelte 5 runes, Tailwind CSS v4 (@tailwindcss/vite), no component library (custom). Icons: emoji/unicode glyphs (lucide-svelte installed but unused). Currency: PHP via Intl en-PH.

## Shared primitives (full source)
- SummaryCard — KPI tile (title, value, description, icon, variant default|positive|negative)
- AccountCard — clickable account tile with icon, type badge, balance
- AccountTypeBadge — pill badge per account type
- AccountBalance — balance block
- LedgerTable — ledger entries table
- RecentTransactions — list card of recent transactions

### `src/lib/components/dashboard/SummaryCard.svelte`

```svelte
<script lang="ts">
    type Props = {
        title: string;
        value: string;
        description?: string;
        icon?: string;
        variant?: "default" | "positive" | "negative";
    };

    let {
        title,
        value,
        description,
        icon,
        variant = "default"
    }: Props = $props();


    const valueColor = {
        default: "text-white",
        positive: "text-emerald-400",
        negative: "text-red-400"
    }[variant];
</script>


<div
        class="
		group
		rounded-2xl
		border
		border-gray-800
		bg-gray-900/80
		p-6
		shadow-xl
		backdrop-blur
		transition
		duration-200
		hover:border-gray-700
		hover:bg-gray-900
	"
>

    <div class="flex items-start justify-between">

        <div>

            <p class="text-sm font-medium text-gray-400">
                {title}
            </p>


            <h2
                    class="
					mt-3
					text-3xl
					font-bold
					tracking-tight
					{valueColor}
				"
            >
                {value}
            </h2>


            {#if description}
                <p class="
					mt-3
					text-xs
					text-gray-500
				">
                    {description}
                </p>
            {/if}

        </div>


        {#if icon}
            <div
                    class="
					flex
					h-12
					w-12
					items-center
					justify-center
					rounded-xl
					bg-gray-800
					text-2xl
					shadow-inner
				"
            >
                {icon}
            </div>
        {/if}

    </div>

</div>
```

### `src/lib/components/dashboard/MetricCard.svelte`

```svelte
<script lang="ts">
    import currency from "$lib/utils/currency.ts";
    interface Props {
        title:string;
        value:number;
        type?: "positive" | "negative";
    }


    let {
        title,
        value,
        type="positive"
    }: Props = $props();


</script>


<div class="rounded-xl border p-6 shadow-sm">

    <p class="text-sm text-gray-500">
        {title}
    </p>


    <p
            class={
			type === "negative"
			? "mt-3 text-3xl font-bold text-red-600"
			: "mt-3 text-3xl font-bold text-green-600"
		}
    >
        {currency.format(value)}
    </p>

</div>
```

### `src/lib/components/accounts/AccountCard.svelte`

```svelte
<script lang="ts">
    import AccountTypeBadge from './AccountTypeBadge.svelte';
    import {currency} from "$lib/utils/currency.ts";
    let { account } = $props();



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
                    <h3 class="font-semibold text-lg text-white">
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
```

### `src/lib/components/accounts/AccountTypeBadge.svelte`

```svelte
<script lang="ts">
    let { type } = $props();

    const colors: Record<string, string> = {
        ASSET: 'bg-green-100 text-green-700',
        LIABILITY: 'bg-red-100 text-red-700',
        CREDIT_CARD: 'bg-blue-100 text-blue-700',
        INCOME: 'bg-emerald-100 text-emerald-700',
        EXPENSE: 'bg-orange-100 text-orange-700',
        RECEIVABLE: 'bg-purple-100 text-purple-700',
        PAYABLE: 'bg-yellow-100 text-yellow-700'
    };
</script>

<span
        class={`inline-block rounded-full px-2 py-1 text-xs font-medium ${
		colors[type] ?? 'bg-gray-100 text-gray-700'
	}`}
>
	{type.replaceAll('_', ' ')}
</span>
```

### `src/lib/components/accounts/AccountBalance.svelte`

```svelte
<script lang="ts">
    import {currency} from "$lib/utils/currency.ts";
    interface Props {
        balance: number;
    }

    let { balance }: Props = $props();
</script>

<div class="rounded-lg bg-primary text-primary-content p-5">

    <div class="text-sm">

        Current Balance

    </div>

    <div class="text-3xl font-bold">
        {currency.format(balance)}

    </div>

</div>
```

### `src/lib/components/accounts/LedgerTable.svelte`

```svelte
<script lang="ts">
    import type {LedgerEntryDto} from "$lib/types/ledger";
    import {currency} from "$lib/utils/currency.ts";


    interface Props {
        entries: LedgerEntryDto[];
    }

    let {entries}: Props = $props();
</script>

<div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

    <div class="overflow-x-auto">
        <table class="w-full text-left">

            <thead class="border-b border-slate-800 bg-slate-950/50">
            <tr>
                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Date
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Description
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Account
                </th>

                <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Type
                </th>

                <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-400">
                    Amount
                </th>
            </tr>
            </thead>

            <tbody class="divide-y divide-slate-800">

            {#each entries as entry}

                <tr class="transition hover:bg-slate-800/50">

                    <td class="whitespace-nowrap px-5 py-4 text-sm text-slate-400">
                        {new Date(entry.createdAt).toLocaleDateString()}
                    </td>

                    <td class="px-5 py-4">
                        <p class="font-medium text-white">
                            {entry.description || "No description"}
                        </p>
                    </td>

                    <td class="px-5 py-4">
                            <span class="font-mono text-xs text-slate-400">
                                {entry.accountId}
                            </span>
                    </td>

                    <td class="px-5 py-4">

                        {#if entry.type === "DEBIT"}

                                <span class="inline-flex rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-400">
                                    Debit
                                </span>

                        {:else}

                                <span class="inline-flex rounded-full bg-blue-500/10 px-3 py-1 text-xs font-medium text-blue-400">
                                    Credit
                                </span>

                        {/if}

                    </td>

                    <td
                            class={`whitespace-nowrap px-5 py-4 text-right font-semibold ${
                                entry.type === "DEBIT"
                                    ? "text-emerald-400"
                                    : "text-blue-400"
                            }`}
                    >
                        {currency.format(Number(entry.amount))}
                    </td>

                </tr>

            {:else}

                <tr>
                    <td
                            colspan="5"
                            class="px-5 py-12 text-center text-sm text-slate-500"
                    >
                        No ledger entries found.
                    </td>
                </tr>

            {/each}

            </tbody>

        </table>
    </div>

</div>
```

### `src/lib/components/dashboard/RecentTransactions.svelte`

```svelte
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
```

### `src/lib/utils/currency.ts`

```svelte
export const currency = new Intl.NumberFormat("en-PH", {
    style: "currency",
    currency: "PHP"
});
```

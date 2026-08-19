<script lang="ts">
    let { data } = $props();

    const currency = new Intl.NumberFormat("en-PH", {
        style: "currency",
        currency: "PHP"
    });

    function formatDate(date: string | null | undefined) {
        if (!date) return "—";

        return new Date(date).toLocaleString("en-PH", {
            year: "numeric",
            month: "short",
            day: "numeric",
            hour: "numeric",
            minute: "2-digit"
        });
    }

    function typeClass(type: string) {
        return type === "DEBIT"
            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
            : "bg-blue-500/10 text-blue-400 border-blue-500/20";
    }

    let totalDebit = $derived(
        data.transaction.entries
            ?.filter((entry: any) => entry.type === "DEBIT")
            .reduce((sum: number, entry: any) => sum + Number(entry.amount), 0) ?? 0
    );

    let totalCredit = $derived(
        data.transaction.entries
            ?.filter((entry: any) => entry.type === "CREDIT")
            .reduce((sum: number, entry: any) => sum + Number(entry.amount), 0) ?? 0
    );

    let balanced = $derived(
        Math.abs(totalDebit - totalCredit) < 0.01
    );
</script>

<svelte:head>
    <title>{data.transaction.description} | FinanKal</title>
</svelte:head>

<div class="space-y-6">

    <!-- Header -->

    <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">

        <div>

            <a
                    href="/transactions"
                    class="text-sm text-slate-500 transition hover:text-slate-300"
            >
                ← Transactions
            </a>

            <h1 class="mt-2 text-3xl font-bold tracking-tight text-white">
                {data.transaction.description}
            </h1>

            <p class="mt-1 text-sm text-slate-400">
                Transaction #{data.transaction.id}
            </p>

        </div>

        <div>
            <span
                    class={`inline-flex rounded-full border px-3 py-1.5 text-sm font-medium ${
                    balanced
                        ? "border-emerald-500/20 bg-emerald-500/10 text-emerald-400"
                        : "border-red-500/20 bg-red-500/10 text-red-400"
                }`}
            >
                {balanced ? "✓ Balanced" : "✕ Not Balanced"}
            </span>
        </div>

    </div>


    <!-- Transaction Information -->

    <div class="grid gap-4 md:grid-cols-3">

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Date
            </p>

            <p class="mt-2 font-semibold text-white">
                {formatDate(data.transaction.createdAt)}
            </p>

        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Total Debit
            </p>

            <p class="mt-2 text-2xl font-bold text-emerald-400">
                {currency.format(totalDebit)}
            </p>

        </div>

        <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

            <p class="text-sm text-slate-400">
                Total Credit
            </p>

            <p class="mt-2 text-2xl font-bold text-blue-400">
                {currency.format(totalCredit)}
            </p>

        </div>

    </div>


    <!-- Ledger Entries -->

    <div class="overflow-hidden rounded-2xl border border-slate-800 bg-slate-900">

        <div class="border-b border-slate-800 px-5 py-4">

            <h2 class="font-semibold text-white">
                Ledger Entries
            </h2>

            <p class="mt-1 text-sm text-slate-400">
                Accounts affected by this transaction.
            </p>

        </div>


        <div class="overflow-x-auto">

            <table class="w-full text-left">

                <thead class="border-b border-slate-800 bg-slate-950/50">

                <tr>

                    <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Account
                    </th>

                    <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Type
                    </th>

                    <th class="px-5 py-4 text-right text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Amount
                    </th>

                    <th class="px-5 py-4 text-xs font-semibold uppercase tracking-wider text-slate-500">
                        Date
                    </th>

                </tr>

                </thead>

                <tbody class="divide-y divide-slate-800">

                {#each data.transaction.entries ?? [] as entry}

                    <tr class="transition hover:bg-slate-800/40">

                        <td class="px-5 py-4">

                            <a
                                    href={`/accounts/${entry.accountId}`}
                                    class="font-medium text-white transition hover:text-blue-400"
                            >
                                {entry.accountName ?? entry.accountId}
                            </a>

                        </td>

                        <td class="px-5 py-4">

                                <span
                                        class={`inline-flex rounded-full border px-2.5 py-1 text-xs font-medium ${typeClass(entry.type)}`}
                                >
                                    {entry.type}
                                </span>

                        </td>

                        <td class="px-5 py-4 text-right font-semibold text-white">
                            {currency.format(entry.amount)}
                        </td>



                    </tr>

                {/each}

                </tbody>

            </table>

        </div>

    </div>


    <!-- Actions -->

    <div class="flex justify-between">

        <a
                href="/transactions"
                class="rounded-xl border border-slate-700 bg-slate-900 px-5 py-3 font-medium text-slate-300 transition hover:bg-slate-800 hover:text-white"
        >
            ← Back to Transactions
        </a>

    </div>

</div>
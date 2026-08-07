<script lang="ts">
    interface Entry {
        accountId: string;
        amount: number;
        type: "DEBIT" | "CREDIT";
    }

    interface Props {
        entries: Entry[];
    }

    let { entries }: Props = $props();

    const totalDebit = $derived(
        entries
            .filter((entry) => entry.type === "DEBIT")
            .reduce((sum, entry) => sum + Number(entry.amount || 0), 0)
    );

    const totalCredit = $derived(
        entries
            .filter((entry) => entry.type === "CREDIT")
            .reduce((sum, entry) => sum + Number(entry.amount || 0), 0)
    );

    const difference = $derived(
        Math.abs(totalDebit - totalCredit)
    );

    const balanced = $derived(
        totalDebit === totalCredit &&
        totalDebit > 0
    );

    const currency = new Intl.NumberFormat("en-PH", {
        style: "currency",
        currency: "PHP"
    });
</script>

<div class="grid gap-4 lg:grid-cols-3">

    <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

        <p class="text-sm text-slate-400">
            Total Debit
        </p>

        <p class="mt-2 text-3xl font-bold text-emerald-400">
            {currency.format(totalDebit)}
        </p>

        <p class="mt-2 text-xs text-slate-500">
            Sum of all debit entries
        </p>

    </div>

    <div class="rounded-2xl border border-slate-800 bg-slate-900 p-5">

        <p class="text-sm text-slate-400">
            Total Credit
        </p>

        <p class="mt-2 text-3xl font-bold text-blue-400">
            {currency.format(totalCredit)}
        </p>

        <p class="mt-2 text-xs text-slate-500">
            Sum of all credit entries
        </p>

    </div>

    <div
            class={`rounded-2xl border p-5 ${
			balanced
				? "border-emerald-700 bg-emerald-950/20"
				: "border-red-700 bg-red-950/20"
		}`}
    >

        <p class="text-sm text-slate-400">
            Transaction Status
        </p>

        {#if balanced}

            <p class="mt-2 text-2xl font-bold text-emerald-400">
                ✓ Balanced
            </p>

            <p class="mt-2 text-sm text-slate-400">
                Debits equal credits.
            </p>

        {:else}

            <p class="mt-2 text-2xl font-bold text-red-400">
                ✕ Not Balanced
            </p>

            <p class="mt-2 text-sm text-slate-400">
                Difference:
                <span class="font-semibold text-red-300">
					{currency.format(difference)}
				</span>
            </p>

        {/if}

    </div>

</div>
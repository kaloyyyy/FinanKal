<script lang="ts">
    interface Account {
        id: string;
        name: string;
        type: string;
    }

    interface Entry {
        accountId: string;
        amount: number;
        type: "DEBIT" | "CREDIT";
    }

    interface Props {
        entry: Entry;
        accounts: Account[];
        onDelete: () => void;
    }
    function update(values: Partial<Entry>) {
        onUpdate({
            ...entry,
            ...values
        });
    }

    let {
        entry,
        accounts=[],
        onDelete,
        onUpdate
    }: Props = $props();

    const amountColor = $derived(
        entry.type === "DEBIT"
            ? "text-emerald-400"
            : "text-red-400"
    );
</script>

<tr class="border-b border-slate-800 last:border-none">

    <td class="px-5 py-4">

        <select
                value={entry.accountId}
                onchange={(e) =>
            update({
                accountId: (e.currentTarget as HTMLSelectElement).value
            })
        }
                class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white outline-none transition focus:border-blue-500"
        >

            <option value="">
                Select Account
            </option>

            <optgroup label="Assets">
                {#each accounts.filter(a => a.type === "ASSET") as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

            <optgroup label="Credit Cards">
                {#each accounts.filter(a => a.type === "CREDIT_CARD") as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

            <optgroup label="Liabilities">
                {#each accounts.filter(a => a.type === "LIABILITY") as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

            <optgroup label="Income">
                {#each accounts.filter(a => a.type === "INCOME") as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

            <optgroup label="Expenses">
                {#each accounts.filter(a => a.type === "EXPENSE") as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

            <optgroup label="Other">
                {#each accounts.filter(a =>
                    ![
                        "ASSET",
                        "LIABILITY",
                        "EXPENSE",
                        "INCOME",
                        "CREDIT_CARD"
                    ].includes(a.type)
                ) as account}
                    <option value={account.id}>
                        {account.name}
                    </option>
                {/each}
            </optgroup>

        </select>

    </td>

    <td class="px-5 py-4">

        <select
                value={entry.type}
                onchange={(e) =>
            update({
                type: (e.currentTarget as HTMLSelectElement)
                    .value as "DEBIT" | "CREDIT"
            })
        }
                class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-white outline-none transition focus:border-blue-500"
        >

            <option value="DEBIT">
                Debit
            </option>

            <option value="CREDIT">
                Credit
            </option>

        </select>

    </td>

    <td class="px-5 py-4">

        <input
                type="number"
                step="0.01"
                min="0"
                value={entry.amount}
                oninput={(e) =>
            update({
                amount: Number(
                    (e.currentTarget as HTMLInputElement).value
                )
            })
        }
                class={`w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-right font-medium outline-none transition focus:border-blue-500 ${amountColor}`}
        />

    </td>

    <td class="px-5 py-4 text-center">

        <button
                type="button"
                onclick={onDelete}
                class="rounded-lg p-2 text-slate-500 transition hover:bg-red-500/10 hover:text-red-400"
        >
            ✕
        </button>

    </td>

</tr>
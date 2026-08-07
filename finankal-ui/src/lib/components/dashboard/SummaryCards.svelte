<script lang="ts">

    import SummaryCard from "./SummaryCard.svelte";


    type Props = {
        totalCredit?: number | string;
        totalDebit?: number | string;
        netWorth?: number | string;
    };


    let {
        totalCredit = 0,
        totalDebit = 0,
        netWorth = 0
    }: Props = $props();



    function currency(
        value: number | string | null | undefined
    ) {

        const amount = Number(value ?? 0);

        return new Intl.NumberFormat(
            "en-PH",
            {
                style: "currency",
                currency: "PHP",
                maximumFractionDigits: 2
            }
        ).format(
            Number.isNaN(amount) ? 0 : amount
        );
    }

</script>


<div
        class="grid w-full grid-cols-3 gap-4"
>
    <SummaryCard
            title="Liquid Money"
            value={currency(totalDebit)}
            description="Available cash across your accounts"
            icon="💵"
            variant="positive"
    />

    <SummaryCard
            title="Total Debt"
            value={currency(totalCredit)}
            description="Outstanding loans and credit"
            icon="💳"
            variant="negative"
    />

    <SummaryCard
            title="Net Worth"
            value={currency(netWorth)}
            description="Assets minus liabilities"
            icon={netWorth >= 0 ? "📈" : "📉"}
            variant={netWorth >= 0 ? "positive" : "negative"}
    />
</div>
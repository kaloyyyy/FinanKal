<script lang="ts">

    import SummaryCard from "./SummaryCard.svelte";
    import { Wallet, CreditCard, TrendingUp, TrendingDown } from "lucide-svelte";


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
        class="grid w-full grid-cols-1 gap-4 md:grid-cols-3"
>
    <SummaryCard
            title="Liquid Assets"
            value={currency(totalDebit)}
            description="Available cash across your accounts"
            icon={Wallet}
            variant="positive"
    />

    <SummaryCard
            title="Total Liabilities"
            value={currency(totalCredit)}
            description="Outstanding loans and credit"
            icon={CreditCard}
            variant="negative"
    />

    <SummaryCard
            title="Net Worth"
            value={currency(netWorth)}
            description="Assets minus liabilities"
            icon={Number(netWorth) >= 0 ? TrendingUp : TrendingDown}
            variant={Number(netWorth) >= 0 ? "positive" : "negative"}
    />
</div>
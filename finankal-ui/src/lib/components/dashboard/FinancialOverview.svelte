<script lang="ts">
    import MetricCard
        from "./MetricCard.svelte";


    interface Props {
        totalCredit: number;
        totalDebit: number;
        netWorth: number;
    }


    let {
        totalCredit,
        totalDebit,
        netWorth
    }: Props = $props();


</script>


<div class="space-y-6">

    <h1 class="text-3xl font-bold">
        Financial Overview
    </h1>


    <div class="grid gap-4 md:grid-cols-3">

        <MetricCard
                title="Net Worth"
                value={netWorth}
                type="positive"
        />


        <MetricCard
                title="Total Credit"
                value={totalCredit}
                type="negative"
        />


        <MetricCard
                title="Total Debit"
                value={totalDebit}
                type="negative"
        />

    </div>

</div>
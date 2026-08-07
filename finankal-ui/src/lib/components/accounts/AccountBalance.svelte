<script lang="ts">
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

        ₱ {balance.toLocaleString()}

    </div>

</div>
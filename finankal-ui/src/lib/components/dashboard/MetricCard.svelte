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
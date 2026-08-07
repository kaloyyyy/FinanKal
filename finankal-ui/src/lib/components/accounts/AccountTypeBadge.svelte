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
<script lang="ts">
    let { type } = $props();

    const colors: Record<string, string> = {
        ASSET: 'bg-emerald-500/10 text-emerald-400',
        LIABILITY: 'bg-red-500/10 text-red-400',
        CREDIT_CARD: 'bg-blue-500/10 text-blue-400',
        INCOME: 'bg-emerald-500/10 text-emerald-400',
        EXPENSE: 'bg-orange-500/10 text-orange-400',
        RECEIVABLE: 'bg-purple-500/10 text-purple-400',
        PAYABLE: 'bg-amber-500/10 text-amber-400'
    };
</script>

<span
        class={`inline-block rounded-full px-2 py-1 text-xs font-medium ${
		colors[type] ?? 'bg-slate-800 text-slate-300'
	}`}
>
	{type.replaceAll('_', ' ')}
</span>
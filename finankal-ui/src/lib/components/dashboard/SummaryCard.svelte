<script lang="ts">
    type Props = {
        title: string;
        value: string;
        description?: string;
        icon?: string;
        variant?: "default" | "positive" | "negative";
    };

    let {
        title,
        value,
        description,
        icon,
        variant = "default"
    }: Props = $props();


    const valueColor = {
        default: "text-white",
        positive: "text-emerald-400",
        negative: "text-red-400"
    }[variant];
</script>


<div
        class="
		group
		rounded-2xl
		border
		border-gray-800
		bg-gray-900/80
		p-6
		shadow-xl
		backdrop-blur
		transition
		duration-200
		hover:border-gray-700
		hover:bg-gray-900
	"
>

    <div class="flex items-start justify-between">

        <div>

            <p class="text-sm font-medium text-gray-400">
                {title}
            </p>


            <h2
                    class="
					mt-3
					text-3xl
					font-bold
					tracking-tight
					{valueColor}
				"
            >
                {value}
            </h2>


            {#if description}
                <p class="
					mt-3
					text-xs
					text-gray-500
				">
                    {description}
                </p>
            {/if}

        </div>


        {#if icon}
            <div
                    class="
					flex
					h-12
					w-12
					items-center
					justify-center
					rounded-xl
					bg-gray-800
					text-2xl
					shadow-inner
				"
            >
                {icon}
            </div>
        {/if}

    </div>

</div>
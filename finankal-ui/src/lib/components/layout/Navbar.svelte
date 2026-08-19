<script lang="ts">
    import { page } from "$app/state";

    let mobileOpen = $state(false);

    const navItems = [
        {
            label: "Dashboard",
            href: "/",
            icon: "▦"
        },
        {
            label: "Accounts",
            href: "/accounts",
            icon: "◉"
        },
        {
            label: "Transactions",
            href: "/transactions",
            icon: "↔"
        },
        {
            label: "Credit Cards",
            href: "/credit-cards",
            icon: "▣"
        }
    ];


    function isActive(href: string) {
        const pathname = page.url.pathname;

        if (href === "/") {
            return pathname === "/";
        }

        return pathname === href || pathname.startsWith(`${href}/`);
    }
</script>

<nav class="border-b border-slate-800 bg-slate-950">
    <div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">

        <!-- Brand -->
        <a
                href="/"
                class="flex items-center gap-3"
                onclick={() => (mobileOpen = false)}
        >
            <div
                    class="flex h-9 w-9 items-center justify-center rounded-xl bg-blue-600 font-bold text-white shadow-lg shadow-blue-600/20"
            >
                F
            </div>

            <div class="hidden sm:block">
                <span class="text-lg font-bold tracking-tight text-white">
                    FinanKal
                </span>
            </div>
        </a>

        <!-- Desktop Navigation -->
        <div class="hidden items-center gap-1 md:flex">
            {#each navItems as item}
                <a
                        href={item.href}
                        class={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition ${
                        isActive(item.href)
                            ? "bg-slate-800 text-white"
                            : "text-slate-400 hover:bg-slate-900 hover:text-white"
                    }`}
                >
                    <span class="text-base">
                        {item.icon}
                    </span>

                    {item.label}
                </a>
            {/each}
        </div>

        <!-- Right Side -->
        <div class="hidden items-center gap-3 md:flex">

            <a
                    href="/transactions/new"
                    class="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-blue-500"
            >
                <span class="text-lg leading-none">+</span>
                New Transaction
            </a>

            <button
                    type="button"
                    class="flex h-9 w-9 items-center justify-center rounded-full border border-slate-700 bg-slate-900 text-sm font-semibold text-slate-300 transition hover:border-slate-600 hover:bg-slate-800 hover:text-white"
                    aria-label="User menu"
            >
                U
            </button>

        </div>

        <!-- Mobile Menu Button -->
        <button
                type="button"
                class="flex h-10 w-10 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 text-slate-300 transition hover:bg-slate-800 hover:text-white md:hidden"
                onclick={() => (mobileOpen = !mobileOpen)}
                aria-label="Toggle navigation"
        >
            {#if mobileOpen}
                ✕
            {:else}
                ☰
            {/if}
        </button>

    </div>

    <!-- Mobile Navigation -->
    {#if mobileOpen}
        <div class="border-t border-slate-800 bg-slate-950 px-4 py-4 md:hidden">

            <div class="space-y-1">

                {#each navItems as item}
                    <a
                            href={item.href}
                            onclick={() => (mobileOpen = false)}
                            class={`flex items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium transition ${
                            isActive(item.href)
                                ? "bg-slate-800 text-white"
                                : "text-slate-400 hover:bg-slate-900 hover:text-white"
                        }`}
                    >
                        <span class="w-5 text-center">
                            {item.icon}
                        </span>

                        {item.label}
                    </a>
                {/each}

            </div>

            <div class="mt-4 border-t border-slate-800 pt-4">

                <a
                        href="/transactions/new"
                        onclick={() => (mobileOpen = false)}
                        class="flex items-center justify-center gap-2 rounded-xl bg-blue-600 px-4 py-3 text-sm font-semibold text-white transition hover:bg-blue-500"
                >
                    <span class="text-lg leading-none">+</span>
                    New Transaction
                </a>

            </div>

        </div>
    {/if}
</nav>
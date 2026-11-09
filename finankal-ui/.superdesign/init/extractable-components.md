# Extractable components

## Layout
## NavBar
- Source: `src/lib/components/layout/Navbar.svelte`
- Category: layout
- Description: Top nav with brand mark, 4 nav links, New Transaction CTA, user avatar
- Extractable props: activeItem (string: dashboard|accounts|transactions|credit-cards, default "dashboard")
- Hardcoded: brand "F" mark + "FinanKal", link labels/icons, CTA text, all CSS

## Basic
## SummaryCard
- Source: `src/lib/components/dashboard/SummaryCard.svelte`
- Category: basic
- Description: KPI tile with label, big value, description, icon
- Extractable props: none (content-driven)
- Hardcoded: styles

## AccountCard
- Source: `src/lib/components/accounts/AccountCard.svelte`
- Category: basic
- Description: Account tile with icon, name, type badge, balance, footer link
- Extractable props: none
- Hardcoded: styles

## AccountTypeBadge
- Source: `src/lib/components/accounts/AccountTypeBadge.svelte`
- Category: basic
- Description: Pill badge per account type
- Extractable props: none

<script lang="ts">
    import { goto } from "$app/navigation";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import { Separator } from "$lib/components/ui/separator";
    import {
        ChevronLeft,
        ShieldCheck,
        Mail as MailIcon,
        Laptop,
        Bug,
        Globe,
        Lock,
    } from "@lucide/svelte";
    import * as m from "$lib/paraglide/messages";
</script>

<div class="min-h-[calc(100vh-1rem)] from-background to-muted/30">
    <div
        class="mx-auto flex max-w-3xl flex-col gap-4 px-4 py-6 sm:px-6 sm:py-10 opacity-80"
    >
        <header class="flex items-start justify-between gap-3">
            <div class="min-w-0">
                <h1
                    class="text-balance text-2xl font-semibold tracking-tight sm:text-3xl"
                >
                    {m.privacy_title()}
                </h1>
                <p class="mt-2 text-sm text-muted-foreground">
                    {m.privacy_description()}
                </p>
            </div>
            <Button
                class="cursor-pointer hover:cursor-pointer"
                variant="ghost"
                onclick={() => goto("/")}
                ><ChevronLeft class="size-4" /> {m.settings_back()}</Button
            >
        </header>

        <!-- Overview Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <ShieldCheck class="size-5" />
                    {m.privacy_overview_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_overview_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Email Content Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <MailIcon class="size-5" />
                    {m.privacy_email_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_email_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Settings & Preferences Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Lock class="size-5" />
                    {m.privacy_settings_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_settings_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Device & Diagnostic Information Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Laptop class="size-5" />
                    {m.privacy_device_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_device_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Bug Reports Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Bug class="size-5" />
                    {m.privacy_bugreport_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_bugreport_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Third-Party Services Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <Globe class="size-5" />
                    {m.privacy_thirdparty_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_thirdparty_text()}
                </p>
            </Card.Content>
        </Card.Root>

        <!-- Data Retention & Questions Card -->
        <Card.Root>
            <Card.Header class="space-y-1">
                <Card.Title class="flex items-center gap-2">
                    <ShieldCheck class="size-5" />
                    {m.privacy_retention_title()}
                </Card.Title>
            </Card.Header>
            <Card.Content>
                <p class="text-sm text-muted-foreground">
                    {m.privacy_retention_text()}
                </p>
                <Separator class="my-4" />
                <p class="text-xs text-muted-foreground text-center">
                    © 2025-{new Date().getFullYear()} 3gIT. {m.credits_copyright()}
                </p>
            </Card.Content>
        </Card.Root>
    </div>
</div>

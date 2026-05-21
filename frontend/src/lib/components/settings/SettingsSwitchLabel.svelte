<script lang="ts">
  import { Switch } from "$lib/components/ui/switch";
  import { Label } from "$lib/components/ui/label";

  type SwitchLabelType = "normal" | "danger";

  // Props
  let {
    featureBool = $bindable(false),
    labelText = null,
    hintText = null,
    infoText = null,
    type = "normal",
    runningInDevMode = false,
  }: {
    featureBool?: boolean;
    labelText: string | null;
    hintText?: string | null;
    infoText: string | null;
    type?: SwitchLabelType;
    runningInDevMode?: boolean;
  } = $props();
</script>

{#if type === "danger"}
  <div
    class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4 border-destructive/30"
  >
    <div class="space-y-1">
      <Label class="text-sm">{labelText}</Label>
      <div class="text-sm text-muted-foreground">
        {hintText}
      </div>
    </div>
    <Switch
      bind:checked={featureBool}
      class="cursor-pointer hover:cursor-pointer"
      disabled={runningInDevMode}
    />
  </div>
  <div class="text-xs text-muted-foreground">
    <strong>{infoText}</strong>
  </div>
{:else}
  <div class="space-y-3">
    <div
      class="flex items-center justify-between gap-4 rounded-lg border bg-card p-4"
    >
      <div>
        <div class="font-medium">
          {labelText}
        </div>
        <div class="text-sm text-muted-foreground">
          {hintText}
        </div>
      </div>
      <Switch
        bind:checked={featureBool}
        class="cursor-pointer hover:cursor-pointer"
      />
    </div>
    <p class="text-xs text-muted-foreground mt-2">
      {infoText}
    </p>
  </div>
{/if}

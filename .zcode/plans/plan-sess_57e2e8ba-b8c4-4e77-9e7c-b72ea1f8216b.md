### Plan: Fix Spec Wizard Select Loading

The goal is to ensure that when editing a Spec Wizard entity, all `ThemedSelect` components correctly display the values loaded from the database, and that cascading options (Architecture, Stacks, etc.) are correctly populated and selected.

#### 1. Analysis of Current implementation
- `SpecWizardDialog.svelte` uses `$effect` to load plugin options when `expert_language_plugin` changes.
- The `ThemedSelect` component uses a native `<select>` element with `bind:value`.
- **Potential Issue:** When the dialog opens, `formData` is set immediately, but the `options` arrays for the selects are initially empty. The native `<select>` element may reset the bound value to `""` if the current value is not present in its options list at the time of rendering.

#### 2. Proposed Changes in `frontend/src/lib/components/ui/Select.svelte`
- Update the logic to ensure that if a `value` is provided but not present in the `options` list, it is still preserved as a hidden or temporary option until the real options load. This is a common pattern for cascading selects.

#### 3. Proposed Changes in `frontend/src/lib/components/settings/SpecWizardDialog.svelte`

**a) De-duplicate loading logic:**
- Remove the redundant `loadPluginOptions(v)` call from the `onValueChange` handler of the Expert Language Plugin. Keep only the reactive `$effect`.

**b) Robust Initialization:**
- In the initialization `$effect`, ensure that `loadExpertPlugins()` and initial `loadPluginOptions()` are awaited or handled in a sequence that prevents the `formData` values from being cleared by empty selects.

#### 4. Verification
- Build the frontend.
- Open an existing Spec Wizard entity for editing.
- Verify that "Expert Language Plugin", "Base Architecture", "Persistence Strategy", and "State Management" all show the correct saved labels immediately.
- Verify that changing the "Expert Language Plugin" correctly updates the options in subsequent steps.
- Ensure no duplicate network calls are made to the `/options` endpoint.
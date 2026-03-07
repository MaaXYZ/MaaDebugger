<template>
    <UCard size="xl">
        <template #header>
            <span class="font-bold">About MaaDebugger</span>
        </template>

        <template #default>
            <div class="flex flex-col gap-3">
                <div class="flex items-center gap-1">
                    <span>MaaDebugger Version: 0.0.1</span>
                    <UTooltip text="Open on GitHub">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaDebugger"
                            target="_blank" icon="i-simple-icons:github" aria-label="open-in-GitHub" />
                    </UTooltip>
                    <UTooltip text="Open on npm">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaDebugger"
                            target="_blank" icon="i-simple-icons:npm" aria-label="open-in-npm" />
                    </UTooltip>
                    <UTooltip text="Open on PyPI">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaDebugger"
                            target="_blank" icon="i-simple-icons:pypi" aria-label="open-in-pypi" />
                    </UTooltip>
                </div>
                <div class="flex items-center gap-2">
                    <span>MaaFramework Version: {{ maaVersion }}</span>
                    <UTooltip text="Open on GitHub">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaFramework"
                            target="_blank" icon="i-simple-icons:github" aria-label="open-in-GitHub" />
                    </UTooltip>
                </div>

                <UButton label="Check for Updates" block />
            </div>
        </template>
    </UCard>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { getMaaFrameworkVersion } from '@/api/http';

const maaVersion = ref("")

onMounted(async () => {
    maaVersion.value = await getMaaFrameworkVersion()
})
</script>

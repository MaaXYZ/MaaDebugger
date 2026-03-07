<template>
    <UCard size="xl">
        <template #header>
            <span class="font-bold">About MaaDebugger</span>
        </template>

        <template #default>
            <div class="flex flex-col gap-4">
                <div class="flex items-center gap-2">
                    <div class="flex items-center gap-1">
                        <span>MaaDebugger Version: {{ maaDebuggerVersion }}</span>
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
                    <span>Build Time: {{ buildTime }}</span>

                    <ULink :to="`https://github.com/MaaXYZ/MaaDebugger/commit/${commitSHA}`" target="_blank"
                        v-if="commitSHA != 'dev'">
                        <span>Commit SHA: {{ commitSHA }}</span>
                    </ULink>
                </div>
                <div class="flex items-center gap-1">
                    <span>MaaFramework Version: {{ maaVersion }}</span>
                    <UTooltip text="Open on GitHub">
                        <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaFramework"
                            target="_blank" icon="i-simple-icons:github" aria-label="open-in-GitHub" />
                    </UTooltip>
                </div>
                <div class="flex items-center gap-1">
                    <span>Channel: {{ currentChannelLabel }}</span>
                    <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaFramework/releases"
                        target="_blank" icon="i-simple-icons:github" aria-label="open-in-GitHub"
                        v-if="currentChannel == GITHUB" />
                    <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaFramework/releases"
                        target="_blank" icon="i-simple-icons:npm" aria-label="open-in-GitHub"
                        v-else-if="currentChannel == NPM" />
                    <UButton color="neutral" variant="ghost" to="https://github.com/MaaXYZ/MaaFramework/releases"
                        target="_blank" icon="i-simple-icons:pypi" aria-label="open-in-GitHub"
                        v-else-if="currentChannel == PYPI" />
                </div>

                <UButton label="Check for Updates" block />
            </div>
        </template>
    </UCard>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { getMaaFrameworkVersion, getChannel, getMaaDebuggerInfos } from '@/api/http';

const GITHUB = "github"
const NPM = "npm"
const PYPI = "pypi"

const maaVersion = ref("")
const maaDebuggerVersion = ref("")
const commitSHA = ref("")
const buildTime = ref("")
const currentChannel = ref("")
const currentChannelLabel = ref("")



onMounted(async () => {
    maaVersion.value = await getMaaFrameworkVersion()

    const dbgInfos = await getMaaDebuggerInfos()
    maaDebuggerVersion.value = dbgInfos.version ?? "dev"
    commitSHA.value = dbgInfos.commit_sha ?? ""
    buildTime.value = dbgInfos.build_time ?? ""

    const channel = await getChannel()
    switch (channel) {
        case GITHUB:
            currentChannel.value = GITHUB
            currentChannelLabel.value = "Github"
            break
        case NPM:
            currentChannel.value = NPM
            currentChannelLabel.value = "npm"
            break
        case PYPI:
            currentChannel.value = PYPI
            currentChannelLabel.value = "PyPI"
            break
        default:
            currentChannel.value = GITHUB
            currentChannelLabel.value = "Github"
    }
})
</script>

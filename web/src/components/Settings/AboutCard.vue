<template>
    <UCard size="xl">
        <template #header>
            <span class="font-bold">About MaaDebugger</span>
        </template>

        <template #default>
            <div class="flex flex-col gap-4">
                <div class="flex flex-wrap items-centers gap-x-4 gap-y-2">
                    <div class="flex min-w-0 flex-wrap items-center gap-1">
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
                    <div
                        class="flex min-w-0 flex-wrap items-center gap-2 rounded-md border border-muted bg-elevated px-3 py-1 text-sm text-toned">
                        <span class="font-medium text-default">Build Info</span>
                        <span class="truncate">Build Time: {{ buildTime }}</span>
                        <ULink v-if="commitSHA != 'dev'"
                            :to="`https://github.com/MaaXYZ/MaaDebugger/commit/${commitSHA}`" target="_blank">
                            <span class="truncate">Commit SHA: {{ commitSHA }}</span>
                        </ULink>
                    </div>
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
                    <UButton v-if="currentChannel == GITHUB" color="neutral" variant="ghost"
                        to="https://github.com/MaaXYZ/MaaFramework/releases" target="_blank"
                        icon="i-simple-icons:github" aria-label="open-in-GitHub" />
                    <UButton v-else-if="currentChannel == NPM" color="neutral" variant="ghost"
                        to="https://github.com/MaaXYZ/MaaFramework/releases" target="_blank" icon="i-simple-icons:npm"
                        aria-label="open-in-GitHub" />
                    <UButton v-else-if="currentChannel == PYPI" color="neutral" variant="ghost"
                        to="https://github.com/MaaXYZ/MaaFramework/releases" target="_blank" icon="i-simple-icons:pypi"
                        aria-label="open-in-GitHub" />
                </div>

                <div class="flex flex-col gap-2">
                    <div
                        class="flex items-center justify-between gap-3 rounded-md border border-muted bg-elevated px-3 py-2">
                        <div class="flex flex-col gap-1">
                            <span class="text-sm font-medium">Include pre-release updates</span>
                            <span class="text-xs text-dimmed">When enabled, update checks may return newer pre-release
                                versions.</span>
                        </div>
                        <USwitch :model-value="updateSettingsStore.showPreRelease"
                            @update:model-value="updateSettingsStore.setShowPreRelease(Boolean($event))" />
                    </div>

                    <UButton label="Check for Updates" :loading="checking" block @click="handleCheckUpdate" />

                    <UAlert v-if="updateResult" :color="updateResult.has_update ? 'info' : 'success'"
                        :icon="updateResult.has_update ? 'i-lucide-download' : 'i-lucide-check-circle'"
                        :title="updateResult.has_update ? 'Update Available' : 'Up to Date'"
                        :description="updateDescription" variant="subtle" />

                    <UAlert v-if="updateError" color="error" icon="i-lucide-alert-circle" title="Check Failed"
                        :description="updateError" variant="subtle" />
                </div>
            </div>
        </template>
    </UCard>
</template>

<script lang="ts" setup>
import { onMounted, ref, computed } from 'vue';
import { getMaaFrameworkVersion, getChannel, getMaaDebuggerInfos, checkForUpdates } from '@/api/http';
import type { UpdateCheckResult } from '@/api/http';
import { useUpdateSettingsStore } from '@/stores/updateSettings';

const GITHUB = "github"
const NPM = "npm"
const PYPI = "pypi"

const maaVersion = ref("")
const maaDebuggerVersion = ref("")
const commitSHA = ref("")
const buildTime = ref("")
const currentChannel = ref("")
const currentChannelLabel = ref("")
const updateSettingsStore = useUpdateSettingsStore()

const checking = ref(false)
const updateResult = ref<UpdateCheckResult | null>(null)
const updateError = ref("")

const updateDescription = computed(() => {
    if (!updateResult.value) return ""
    if (updateResult.value.has_update) {
        const channel = updateResult.value.track || (updateResult.value.nightly ? "nightly" : "release")
        let desc = `Current: ${updateResult.value.current_version} → Latest: ${updateResult.value.latest_version} (${channel})`
        if (updateResult.value.note) {
            desc += `\n${updateResult.value.note}`
        }
        return desc
    }
    return `You are running the latest version (${updateResult.value.current_version})`
})

async function handleCheckUpdate() {
    checking.value = true
    updateError.value = ""
    updateResult.value = null

    if (maaDebuggerVersion.value === "dev") {
        updateError.value = "You are running a development build. Update checking is not available."
        checking.value = false
        return
    }

    try {
        const result = await checkForUpdates(updateSettingsStore.showPreRelease)
        updateResult.value = result
    } catch (e) {
        updateError.value = e instanceof Error ? e.message : "Unknown error"
    } finally {
        checking.value = false
    }
}

function formatBuildTime(timestamp: string | null | undefined) {
    if (!timestamp) {
        return ""
    }

    const unixTimestamp = Number(timestamp)
    if (!Number.isFinite(unixTimestamp)) {
        return timestamp
    }

    const date = new Date(unixTimestamp * 1000)
    if (Number.isNaN(date.getTime())) {
        return timestamp
    }

    return new Intl.DateTimeFormat(undefined, {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
    }).format(date)
}

onMounted(async () => {
    const [_maaVersion, _maaDebuggerVersion, _channel] = await Promise.all([
        getMaaFrameworkVersion(),
        getMaaDebuggerInfos(),
        getChannel()
    ])

    maaVersion.value = _maaVersion

    maaDebuggerVersion.value = _maaDebuggerVersion.version ?? "dev"
    commitSHA.value = _maaDebuggerVersion.commit_sha ?? ""
    buildTime.value = formatBuildTime(_maaDebuggerVersion.build_time)

    switch (_channel) {
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

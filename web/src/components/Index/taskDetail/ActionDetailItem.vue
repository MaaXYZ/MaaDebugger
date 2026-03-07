<template>
    <div class="flex flex-col gap-2 rounded-lg border border-default p-3">
        <!-- Header: action type + success/fail -->
        <div class="flex flex-row items-center gap-2 flex-wrap">
            <UBadge :color="detail.success ? 'success' : 'error'" variant="subtle">
                {{ detail.success ? 'Success' : 'Failed' }}
            </UBadge>
            <UBadge color="info" variant="subtle">{{ detail.action }}</UBadge>
            <span class="text-sm font-medium">{{ detail.name }}</span>
        </div>

        <!-- Box -->
        <div v-if="detail.box" class="text-xs text-dimmed">
            Box: [{{ detail.box.x }}, {{ detail.box.y }}, {{ detail.box.w }}, {{ detail.box.h }}]
        </div>

        <!-- Action Result Details -->
        <div v-if="detail.result" class="flex flex-col gap-1.5">
            <!-- Click -->
            <template v-if="detail.result.type === 'Click'">
                <InfoRow label="Point" :value="formatPoint((detail.result as ClickActionResult).point)" />
            </template>

            <!-- LongPress -->
            <template v-else-if="detail.result.type === 'LongPress'">
                <InfoRow label="Point" :value="formatPoint((detail.result as LongPressActionResult).point)" />
                <InfoRow label="Duration" :value="`${(detail.result as LongPressActionResult).duration}ms`" />
            </template>

            <!-- Swipe -->
            <template v-else-if="detail.result.type === 'Swipe'">
                <InfoRow label="Begin" :value="formatPoint((detail.result as SwipeActionResult).begin)" />
                <InfoRow label="End" :value="(detail.result as SwipeActionResult).end.map(formatPoint).join(' → ')" />
                <InfoRow label="Duration"
                         :value="(detail.result as SwipeActionResult).duration.map(d => `${d}ms`).join(', ')" />
            </template>

            <!-- MultiSwipe -->
            <template v-else-if="detail.result.type === 'MultiSwipe'">
                <div v-for="(swipe, idx) in (detail.result as MultiSwipeActionResult).swipes" :key="idx"
                     class="flex flex-col gap-1 rounded border border-default p-2">
                    <span class="text-xs text-dimmed font-medium">Swipe #{{ idx }}</span>
                    <InfoRow label="Begin" :value="formatPoint(swipe.begin)" />
                    <InfoRow label="End" :value="swipe.end.map(formatPoint).join(' → ')" />
                    <InfoRow label="Duration" :value="swipe.duration.map(d => `${d}ms`).join(', ')" />
                </div>
            </template>

            <!-- TouchDown / TouchMove / TouchUp -->
            <template v-else-if="['TouchDown', 'TouchMove', 'TouchUp'].includes(detail.result.type)">
                <InfoRow label="Point" :value="formatPoint((detail.result as TouchActionResult).point)" />
                <InfoRow label="Contact" :value="String((detail.result as TouchActionResult).contact)" />
            </template>

            <!-- Scroll -->
            <template v-else-if="detail.result.type === 'Scroll'">
                <InfoRow label="Point" :value="formatPoint((detail.result as ScrollActionResult).point)" />
                <InfoRow label="Delta"
                         :value="`dx: ${(detail.result as ScrollActionResult).dx}, dy: ${(detail.result as ScrollActionResult).dy}`" />
            </template>

            <!-- ClickKey / KeyDown / KeyUp -->
            <template v-else-if="['ClickKey', 'KeyDown', 'KeyUp'].includes(detail.result.type)">
                <InfoRow label="Keycode" :value="(detail.result as ClickKeyActionResult).keycode.join(', ')"
                         :controller-type="detail.controller_type" />
            </template>

            <!-- LongPressKey -->
            <template v-else-if="detail.result.type === 'LongPressKey'">
                <InfoRow label="Keycode" :value="(detail.result as LongPressKeyActionResult).keycode.join(', ')"
                         :controller-type="detail.controller_type" />
                <InfoRow label="Duration" :value="`${(detail.result as LongPressKeyActionResult).duration}ms`"
                         :controller-type="detail.controller_type" />
            </template>

            <!-- InputText -->
            <template v-else-if="detail.result.type === 'InputText'">
                <InfoRow label="Text" :value="(detail.result as InputTextActionResult).text" />
            </template>

            <!-- StartApp / StopApp -->
            <template v-else-if="['StartApp', 'StopApp'].includes(detail.result.type)">
                <InfoRow label="Package" :value="(detail.result as AppActionResult).package" />
            </template>

            <!-- Shell / Command -->
            <template v-else-if="['Shell', 'Command'].includes(detail.result.type)">
                <InfoRow label="Cmd" :value="(detail.result as ShellActionResult).cmd" />
                <InfoRow label="Timeout" :value="`${(detail.result as ShellActionResult).timeout}ms`" />
                <InfoRow label="Success" :value="String((detail.result as ShellActionResult).success)" />
                <div v-if="(detail.result as ShellActionResult).output" class="flex flex-col gap-0.5">
                    <span class="text-xs text-dimmed">Output:</span>
                    <pre
                        class="text-xs bg-muted rounded p-2 overflow-x-auto max-h-32 whitespace-pre-wrap break-all">{{ (detail.result as ShellActionResult).output }}</pre>
                </div>
            </template>

            <!-- DoNothing / StopTask / Custom / unknown -->
            <template v-else>
                <span class="text-xs text-dimmed italic">{{ detail.result.type }}</span>
            </template>
        </div>

        <!-- detail_json fallback -->
        <div v-if="detail.detail_json" class="flex flex-col gap-1.5">
            <div class="flex items-center justify-between gap-2">
                <span class="text-xs text-dimmed font-medium">Detail JSON:</span>
                <UButton :icon="copied ? 'i-lucide-check' : 'i-lucide-copy'" size="xs" variant="ghost"
                         :color="copied ? 'success' : 'neutral'" @click="copyDetailJson(detail.detail_json)" />
            </div>
            <pre
                class="rounded-md border border-default bg-muted p-3 text-xs overflow-x-auto max-h-75 whitespace-pre-wrap break-all">
        <code>{{ formatJson(detail.detail_json) }}</code></pre>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type {
    ActionDetailResponse,
    PointResponse,
    ClickActionResult,
    LongPressActionResult,
    SwipeActionResult,
    MultiSwipeActionResult,
    TouchActionResult,
    ScrollActionResult,
    ClickKeyActionResult,
    LongPressKeyActionResult,
    InputTextActionResult,
    AppActionResult,
    ShellActionResult,
} from './types'
import InfoRow from './ActionInfoRow.vue'

defineProps<{
    detail: ActionDetailResponse
}>()

const copied = ref(false)
let copyTimer: ReturnType<typeof setTimeout> | null = null

async function copyDetailJson(val: unknown) {
    try {
        await navigator.clipboard.writeText(formatJson(val))
        copied.value = true
        if (copyTimer) clearTimeout(copyTimer)
        copyTimer = setTimeout(() => {
            copied.value = false
        }, 2000)
    } catch {
        // noop
    }
}

function formatPoint(p: PointResponse): string {
    return `(${p.x}, ${p.y})`
}

function formatJson(val: unknown): string {
    return JSON.stringify(val, null, 2)
}
</script>

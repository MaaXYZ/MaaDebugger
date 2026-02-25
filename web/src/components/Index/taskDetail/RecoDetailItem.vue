<template>
    <div class="flex flex-col gap-1.5 rounded-lg border border-default p-2">
        <!-- Header -->
        <div class="flex flex-row items-center gap-2 flex-wrap">
            <UBadge :color="detail.hit ? 'success' : 'error'" variant="subtle" size="xs">
                {{ detail.hit ? 'Hit' : 'Miss' }}
            </UBadge>
            <UBadge color="info" variant="subtle" size="xs">{{ detail.algorithm }}</UBadge>
            <span class="text-xs font-medium">{{ detail.name }}</span>
        </div>

        <!-- Box -->
        <div v-if="detail.box" class="text-xs text-dimmed">
            Box: [{{ detail.box.x }}, {{ detail.box.y }}, {{ detail.box.w }}, {{ detail.box.h }}]
        </div>

        <!-- Nested Combined Result (recursive And/Or) -->
        <div v-if="detail.combined_result && detail.combined_result.length > 0 && depth < 10"
            class="flex flex-col gap-1.5 mt-1">
            <span class="text-xs text-dimmed">Combined ({{ detail.algorithm }}):</span>
            <div class="pl-2 border-l-2 border-default flex flex-col gap-1.5">
                <RecoDetailItem v-for="(sub, idx) in detail.combined_result" :key="idx" :detail="sub"
                    :depth="depth + 1" />
            </div>
        </div>

        <!-- Detail JSON (collapsed) -->
        <details v-if="detail.detail_json" class="text-xs">
            <summary class="text-dimmed cursor-pointer hover:text-default transition-colors">Detail JSON</summary>
            <pre class="bg-elevated rounded-lg p-2 overflow-auto max-h-48 mt-1 border border-default font-mono leading-relaxed whitespace-pre-wrap break-words"
                v-html="syntaxHighlight(detail.detail_json)"></pre>
        </details>
    </div>
</template>

<script setup lang="ts">
import type { RecoDetailResponse } from './types'

defineProps<{
    detail: RecoDetailResponse
    depth: number
}>()

function syntaxHighlight(json: unknown): string {
    const str = JSON.stringify(json, null, 2)
    return str.replace(
        /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
        (match) => {
            let cls = 'color: var(--ui-text-muted)' // number
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'color: var(--ui-primary); font-weight: 500' // key
                } else {
                    cls = 'color: var(--ui-success)' // string
                }
            } else if (/true|false/.test(match)) {
                cls = 'color: var(--ui-warning)' // boolean
            } else if (/null/.test(match)) {
                cls = 'color: var(--ui-error)' // null
            }
            return `<span style="${cls}">${match}</span>`
        }
    )
}
</script>

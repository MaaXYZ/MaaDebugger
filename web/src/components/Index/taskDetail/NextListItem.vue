<template>
    <div class="inline-flex max-w-full flex-wrap items-start gap-1.5 align-top min-w-0 w-fit">
        <template v-for="(entry, idx) in groupedEntries" :key="`entry-${idx}`">
            <template v-if="entry.recos.length === 0">
                <UTooltip :text="formatItemLabel(entry.item)">
                    <UButton size="sm" variant="outline" color="neutral" disabled
                             class="max-w-full min-w-0 overflow-hidden">
                        <span class="truncate block min-w-0">{{ formatItemLabel(entry.item) }}</span>
                    </UButton>
                </UTooltip>
            </template>

            <template v-else-if="entry.recos.length === 1 && entry.recos[0]?.msg.name === entry.item.name">
                <RecoButton :reco="entry.recos[0]" :info="entry.item" use-warning
                            @request-detail="$emit('requestDetail', $event)" />
            </template>

            <div v-else
                 class="inline-flex max-w-full flex-wrap items-start gap-1.5 rounded-md border border-default px-2 py-1">
                <div class="inline-flex max-w-full items-center gap-1.5">
                    <UTooltip :text="formatItemLabel(entry.item)">
                        <UButton size="sm" variant="soft" color="neutral" disabled
                                 class="max-w-full min-w-0 overflow-hidden">
                            <span class="truncate block min-w-0">{{ formatItemLabel(entry.item) }}</span>
                        </UButton>
                    </UTooltip>

                    <UBadge v-if="entry.algorithmType" size="sm" color="info" variant="subtle">
                        {{ entry.algorithmType }}
                    </UBadge>
                </div>

                <template v-for="(reco, recoIdx) in entry.recos" :key="`entry-${idx}-reco-${recoIdx}`">
                    <RecoButton :reco="reco" @request-detail="$emit('requestDetail', $event)" />
                </template>
            </div>
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getNodeDetail } from '@/api/http'
import type { NextListScope, NextListItem, RecoScope } from './types'
import RecoButton from './RecoButton.vue'

const props = defineProps<{
    nextList: NextListScope
}>()

defineEmits<{
    requestDetail: [recoId: number]
}>()

interface GroupedNextEntry {
    item: NextListItem
    recos: RecoScope[]
    algorithmType?: 'And' | 'Or'
}

const algorithmMap = ref<Record<string, 'And' | 'Or' | undefined>>({})

const groupedEntries = computed<GroupedNextEntry[]>(() => {
    const list = props.nextList.msg.list ?? []
    const recos = props.nextList.childs ?? []
    const entries: GroupedNextEntry[] = []
    let recoIndex = 0

    for (const item of list) {
        const entry: GroupedNextEntry = {
            item,
            recos: [],
            algorithmType: algorithmMap.value[item.name],
        }

        while (recoIndex < recos.length) {
            const reco = recos[recoIndex]
            if (!reco) break

            entry.recos.push(reco)
            recoIndex += 1

            if (reco.msg.name === item.name) {
                break
            }
        }

        entries.push(entry)
    }

    while (recoIndex < recos.length) {
        const reco = recos[recoIndex]
        if (!reco) break

        entries.push({
            item: {
                name: reco.msg.name,
                jump_back: false,
                anchor: false,
            },
            recos: [reco],
            algorithmType: algorithmMap.value[reco.msg.name],
        })
        recoIndex += 1
    }

    return entries
})

watch(
    groupedEntries,
    async (entries) => {
        const targets = entries
            .filter((entry) => entry.recos.length > 1)
            .map((entry) => entry.item.name)
            .filter((name) => !(name in algorithmMap.value))

        if (targets.length === 0) return

        const updates: Record<string, 'And' | 'Or' | undefined> = {}
        await Promise.all(targets.map(async (name) => {
            try {
                const detail = await getNodeDetail(name)
                const algorithm = detail?.recognition?.algorithm
                if (algorithm === 'And' || algorithm === 'Or') {
                    updates[name] = algorithm
                    return
                }
            } catch {
                // ignore lookup failure, leave badge hidden
            }
            updates[name] = undefined
        }))

        algorithmMap.value = {
            ...algorithmMap.value,
            ...updates,
        }
    },
    { immediate: true },
)

function formatItemLabel(item: NextListItem): string {
    const label = item.label?.trim()
    if (label) return label

    let fallback = item.name
    if (item.anchor) fallback = `[Anchor] ${fallback}`
    if (item.jump_back) fallback = `[JumpBack] ${fallback}`
    return fallback
}
</script>

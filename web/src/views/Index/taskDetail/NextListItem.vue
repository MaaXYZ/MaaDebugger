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

            <template
                v-else-if="entry.recos.length === 1 && entry.recos[0]?.msg.name === entry.item.name && !entry.recos[0]?.childs.length">
                <div class="flex max-w-full min-w-0 flex-col gap-2">
                    <RecoButton :reco="entry.recos[0]" :info="entry.item" :algorithm-type="entry.item.algorithm"
                        use-warning @request-detail="$emit('requestDetail', $event)" />
                    <SubflowDisclosure v-if="entry.recos[0] && entry.recos[0].childs.length > 0"
                        :label="t('taskDetail.internalFlow')" kind="reco" :count="countRecoSubflowNodes(entry.recos[0])">
                        <SubflowTree :nodes="entry.recos[0].childs" kind="reco"
                            @request-detail="$emit('requestDetail', $event)"
                            @request-action-detail="$emit('requestActionDetail', $event)" />
                    </SubflowDisclosure>
                </div>
            </template>

            <div v-else
                class="inline-flex max-w-full flex-col items-start gap-1.5 rounded-md border border-default px-2 py-1">
                <div class="inline-flex max-w-full flex-wrap items-start gap-1.5">
                    <UTooltip :text="formatItemLabel(entry.item)">
                        <UButton size="sm" variant="soft" color="neutral" disabled
                            class="max-w-full min-w-0 overflow-hidden">
                            <span class="truncate block min-w-0">{{ formatItemLabel(entry.item) }}</span>
                        </UButton>
                    </UTooltip>

                    <template v-for="(reco, recoIdx) in entry.recos" :key="`entry-${idx}-reco-${recoIdx}`">
                        <div class="flex max-w-full min-w-0 flex-col gap-2">
                            <RecoButton :reco="reco" :algorithm-type="entry.item.algorithm"
                                @request-detail="$emit('requestDetail', $event)" />
                            <SubflowDisclosure v-if="reco.childs.length > 0" :label="t('taskDetail.internalFlow')" kind="reco"
                                :count="countRecoSubflowNodes(reco)">
                                <SubflowTree :nodes="reco.childs" kind="reco"
                                    @request-detail="$emit('requestDetail', $event)"
                                    @request-action-detail="$emit('requestActionDetail', $event)" />
                            </SubflowDisclosure>
                        </div>
                    </template>
                </div>
            </div>
        </template>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { NextListScope, NextListItem, RecoScope } from '@/types/taskDetail'
import RecoButton from './RecoButton.vue'
import SubflowTree from './SubflowTree.vue'
import SubflowDisclosure from './SubflowDisclosure.vue'
import { countRecoSubflowNodes } from './scopeTree'

const props = defineProps<{
    nextList: NextListScope
}>()

const { t } = useI18n()

defineEmits<{
    requestDetail: [recoId: number]
    requestActionDetail: [actionId: number]
}>()

interface GroupedNextEntry {
    item: NextListItem
    recos: RecoScope[]
}

const groupedEntries = computed<GroupedNextEntry[]>(() => {
    const list = props.nextList.msg.list ?? []
    const recos = props.nextList.childs ?? []
    const entries: GroupedNextEntry[] = []
    let recoIndex = 0

    for (const item of list) {
        const entry: GroupedNextEntry = {
            item,
            recos: [],
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
        })
        recoIndex += 1
    }

    return entries
})


function formatItemLabel(item: NextListItem): string {
    const label = item.label?.trim()
    if (label) return label

    let fallback = item.name
    if (item.anchor) fallback = t('taskDetail.anchorLabel', { name: fallback })
    if (item.jump_back) fallback = t('taskDetail.jumpBackLabel', { name: fallback })
    return fallback
}
</script>

<template>
    <div class="gap-2 flex flex-col">
        <UTabs key="value" v-model="tabsModel" :items="tabs" />
        <div v-show="tabsModel === 'controller'" class="gap-2 flex flex-col">
            <ControllerCard />
        </div>
        <div v-show="tabsModel === 'resource'" class="gap-2 flex flex-col">
            <InterfaceCard />
            <ResourceCard />
        </div>
        <div v-show="tabsModel === 'agent'" class="gap-2 flex flex-col">
            <AgentCard />
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import type { TabsItem } from '@nuxt/ui';

import InterfaceCard from './InterfaceCard.vue';
import ControllerCard from './ControllerCard.vue';
import ResourceCard from './ResourceCard.vue';
import AgentCard from './AgentCard.vue';

const tabsModel = ref("controller")
const tabs: TabsItem[] = [
    { label: 'Controller', value: 'controller' },
    { label: 'Resource', value: 'resource' },
    { label: 'Agent', value: 'agent' }
];
</script>

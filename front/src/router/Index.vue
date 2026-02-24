y<template>
    <div class="flex flex-row items-start w-full">
        <div class="w-full max-w-md">
            <div class="flex flex-col gap-4 sm:gap-6 sm:p-6 w-full">
                <ControllerCard />
                <ResourceCard />
            </div>
        </div>
        <div class="w-full max-w-md">
            <div class="flex flex-col gap-4 sm:gap-6 sm:p-6 w-full">
                <AgentCard />
                <TaskCard />
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import ControllerCard from "@/components/Index/ControllerCard.vue";
import ResourceCard from "@/components/Index/ResourceCard.vue";
import AgentCard from "@/components/Index/AgentCard.vue";
import TaskCard from "@/components/Index/TaskCard.vue";
</script>

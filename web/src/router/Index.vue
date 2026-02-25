<template>
    <div class="flex flex-col lg:grid lg:grid-cols-3 items-start w-full gap-4 p-4 lg:p-6">
        <!-- 第一列: Controller / Resource / Agent -->
        <div class="w-full flex flex-col gap-4 order-4 lg:order-0">
            <ControllerCard />
            <ResourceCard />
            <AgentCard />
            <TaskCard />
        </div>
        <!-- 第二列: Screenshot -->
        <div class="w-full flex flex-col gap-4 order-2 lg:order-0">
            <ScreenshotCard />
        </div>
        <!-- 第三列: Task Detail -->
        <div class="w-full flex flex-col gap-4 order-1 lg:order-0">
            <TaskDetailCard />
        </div>
    </div>
</template>

<script setup lang="ts">
import ControllerCard from "@/components/Index/ControllerCard.vue";
import ResourceCard from "@/components/Index/ResourceCard.vue";
import AgentCard from "@/components/Index/AgentCard.vue";
import TaskCard from "@/components/Index/TaskCard.vue";
import ScreenshotCard from "@/components/Index/ScreenshotCard.vue";
import TaskDetailCard from "@/components/Index/TaskDetailCard.vue";
</script>

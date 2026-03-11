<template>
    <div class="flex flex-col lg:grid lg:grid-cols-3 items-start w-full gap-4 p-4 lg:p-6">
        <!-- 第一列: Interface / Controller / Resource / Agent -->
        <div class="w-full flex flex-col gap-4 order-4 lg:order-0">
            <LeftTabs />
        </div>
        <!-- 第二列: Task (includes Screenshot) -->
        <div class="w-full flex flex-col gap-4 order-2 lg:order-0">
            <TaskCard />

        </div>
        <!-- 第三列: Task Detail -->
        <div class="w-full flex flex-col gap-4 order-1 lg:order-0">
            <TaskDetailCard />
            <!-- FocusCard 占位符 -->
            <UCard>
                <template #header>
                    <h3 class="text-sm font-semibold">Focus</h3>
                </template>
                <p class="text-sm text-muted">Focus placeholder</p>
            </UCard>
        </div>
    </div>
</template>

<script setup lang="ts">
import LeftTabs from "@/components/Index/LeftTabs.vue";
import TaskCard from "@/components/Index/TaskCard.vue";
import TaskDetailCard from "@/components/Index/TaskDetailCard.vue";
</script>

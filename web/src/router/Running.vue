<template>
    <div class="flex flex-col lg:grid lg:grid-cols-2 items-start w-full gap-4 p-4 lg:p-6">
        <!-- 第一列(原第二列): Task (includes Screenshot) -->
        <div class="w-full flex flex-col gap-4 order-2 lg:order-0">
            <TaskCard />
            <!-- FocusCard 占位符 -->
            <UCard>
                <template #header>
                    <h3 class="text-sm font-semibold">Focus</h3>
                </template>
                <p class="text-sm text-muted">Focus placeholder</p>
            </UCard>
        </div>

        <!-- 第二列(原第三列): Task Detail -->
        <div class="w-full min-h-screen flex flex-col gap-4 order-1 lg:order-0">
            <TaskDetailCard />
        </div>
    </div>
</template>

<script setup lang="ts">
import TaskCard from "@/components/Index/TaskCard.vue";
import TaskDetailCard from "@/components/Index/TaskDetailCard.vue";
</script>

<template>
    <div class="w-full h-full flex flex-row items-start gap-4 p-4">
        <div class="w-1/2 h-full flex flex-col gap-4">
            <TaskCard />
        </div>

        <div class="w-1/2 h-full flex flex-col gap-4">
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
import TaskCard from "@/components/Index/TaskCard.vue";
import TaskDetailCard from "@/components/Index/TaskDetailCard.vue";
</script>

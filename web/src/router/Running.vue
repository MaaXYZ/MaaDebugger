<template>
    <div class="w-full min-h-full flex flex-col items-start gap-4 p-2 lg:p-2 lg:grid lg:grid-cols-2">
        <!-- 第一列: Task (includes Screenshot) -->
        <div class="w-full h-full flex flex-col gap-4 order-2 lg:order-0">
            <TaskCard />

        </div>
        <!-- 第二列: Task Detail -->
        <div class="w-full h-full flex flex-col gap-4 order-1 lg:order-0">
            <TaskDetailCard />
        </div>
    </div>
</template>

<script setup lang="ts">
import TaskCard from "@/components/Index/TaskCard.vue";
import TaskDetailCard from "@/components/Index/TaskDetailCard.vue";
</script>

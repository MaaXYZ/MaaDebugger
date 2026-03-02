<template>
    <UContainer class="py-4 lg:py-6">
        <div class="mb-4 flex items-center justify-end">
            <UTooltip text="Back to dashboard">
                <UButton to="/" color="neutral" variant="ghost" icon="i-lucide-arrow-left" aria-label="Back" />
            </UTooltip>
        </div>

        <TaskDetailCard />
    </UContainer>
</template>

<script setup lang="ts">
import TaskDetailCard from '@/components/Index/TaskDetailCard.vue'
</script>

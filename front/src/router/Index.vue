<template>
    <div class="flex flex-col lg:grid lg:grid-cols-3 items-start w-full gap-4 p-4 lg:p-6">
        <!-- 小屏: order-4, 大屏: 第一列第一行 -->
        <div class="w-full order-4 lg:order-0 lg:col-start-1 lg:row-start-1">
            <ControllerCard />
        </div>
        <!-- 小屏: order-3, 大屏: 第一列第二行 -->
        <div class="w-full order-3 lg:order-0 lg:col-start-1 lg:row-start-2">
            <ResourceCard />
        </div>
        <!-- 小屏: order-5, 大屏: 第一列第三行 -->
        <div class="w-full order-5 lg:order-0 lg:col-start-1 lg:row-start-3">
            <AgentCard />
        </div>
        <!-- 小屏: order-2, 大屏: 第二列 -->
        <div class="w-full order-2 lg:order-0 lg:col-start-2 lg:row-start-1 lg:row-span-3">
            <TaskCard />
        </div>
        <!-- 小屏: order-1 (置顶), 大屏: 第三列 -->
        <div class="w-full order-1 lg:order-0 lg:col-start-3 lg:row-start-1 lg:row-span-3">
            <!--ImgCard and Log-->
        </div>
    </div>
</template>

<script lang="ts">
import ControllerCard from "@/components/Index/ControllerCard.vue";
import ResourceCard from "@/components/Index/ResourceCard.vue";
import AgentCard from "@/components/Index/AgentCard.vue";
import TaskCard from "@/components/Index/TaskCard.vue";
</script>

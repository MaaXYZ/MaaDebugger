<script setup lang="ts">
import Index from '@/router/Index.vue';
</script>

<template>
  <UApp :toaster="{ position: 'button-right', progress: false, duration: 8000 }">
    <UMain>
      <Index />
    </UMain>
  </UApp>
</template>

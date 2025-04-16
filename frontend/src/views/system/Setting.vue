<template>
  <div></div>
</template>

<script  setup lang="ts">
import { router } from '@/router';


onMounted(() => {
  router.push('/exception/incomplete')
})
</script>
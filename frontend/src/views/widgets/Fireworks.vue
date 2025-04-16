<template>
  <div class="page-content">
    <div class="action-buttons">
      <el-button :disabled="isLaunching" v-ripple @click="handleSingleLaunch"
        >✨ 放个小烟花</el-button
      >
      <el-button :disabled="isLaunching" v-ripple @click="handleImageLaunch(bp)"
        >🎉 打开幸运红包</el-button
      >
      <el-button :disabled="isLaunching" v-ripple @click="handleMultipleLaunch('')"
        >🎆 璀璨烟火秀</el-button
      >
      <el-button :disabled="isLaunching" v-ripple @click="handleImageLaunch(sd)"
        >❄️ 飘点小雪花</el-button
      >
      <el-button :disabled="isLaunching" v-ripple @click="handleMultipleLaunch(sd)"
        >❄️ 浪漫暴风雪</el-button
      >
    </div>
  </div>
</template>

<script setup lang="ts">
  import mittBus from '@/utils/mittBus'
  import { ref } from 'vue'

  import bp from '@imgs/ceremony/hb.png'
  import sd from '@imgs/ceremony/sd.png'

  const timerRef = ref<ReturnType<typeof setInterval> | null>(null)
  const isLaunching = ref(false)

  const triggerFireworks = (count: number, src: string) => {
    // 清除之前的定时器
    if (timerRef.value) {
      clearInterval(timerRef.value)
      timerRef.value = null
    }

    isLaunching.value = true // 开始发射时设置状态

    let fired = 0
    timerRef.value = setInterval(() => {
      mittBus.emit('triggerFireworks', src)
      fired++

      // 达到指定次数后清除定时器
      if (fired >= count) {
        clearInterval(timerRef.value!)
        timerRef.value = null
        isLaunching.value = false // 发射完成后解除禁用
      }
    }, 1000)
  }

  // 简化后的处理函数
  const handleSingleLaunch = () => {
    mittBus.emit('triggerFireworks')
  }

  const handleMultipleLaunch = (src: string) => {
    triggerFireworks(10, src)
  }

  const handleImageLaunch = (src: string) => {
    mittBus.emit('triggerFireworks', src)
  }

  // 组件卸载时清理定时器
  onUnmounted(() => {
    if (timerRef.value) {
      clearInterval(timerRef.value)
      timerRef.value = null
    }
  })
</script>

<style lang="scss" scoped>
  .action-buttons {
    margin-bottom: 20px;
  }
</style>

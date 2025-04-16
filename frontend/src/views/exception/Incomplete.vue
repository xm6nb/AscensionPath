<template>
  <div class="page-content success">
    <i class="iconfont-sys icon">&#xe892;</i>
    <h1 class="title">抱歉</h1>
    <p class="msg">该页面尚未完成</p>
    <p class="msg">敬请期待更精彩的呈现！</p>
    <div class="res">
      <button class="btn" @click="toHome">↩️ 返回首页</button>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { HOME_PAGE } from '@/router'
  const router = useRouter()

  const toHome = () => {
    router.push(HOME_PAGE)
  }
</script>

<style lang="scss" scoped>
  .success {
    box-sizing: border-box;
    padding: 15px 100px !important;
    text-align: center;

    .icon {
      display: block;
      margin-top: 6vh;
      font-size: 80px;
      color: #19be6b !important;
    }

    .title {
      margin-top: 20px;
      font-size: 30px;
      font-weight: 500;
      color: var(--art-text-gray-900) !important;
    }

    .msg {
      margin-top: 20px;
      font-size: 16px;
      color: #808695;
    }

    .res {
      padding: 22px 30px;
      margin-top: 30px;
      text-align: left;
      background-color: #f8f8f9;
      border-radius: 5px;

      p {
        padding: 8px 0;
        font-size: 15px;
        color: #808695;
      }
    }

    .btn-group {
      margin-top: 50px;
    }
  }

  .dark {
    .success {
      .res {
        background: #28282a;
      }
    }
  }

  @media screen and (max-width: $device-phone) {
    .success {
      padding: 15px 25px !important;

      .icon {
        margin-top: 4vh;
        font-size: 60px;
      }

      .title {
        margin-top: 10px;
        font-size: 25px;
      }

      .res {
        padding: 10px 30px;
      }
    }
  }

  /* From Uiverse.io by SujitAdroja */
  .btn {
    color: purple;
    text-transform: uppercase;
    text-decoration: none;
    border: 2px solid purple;
    padding: 10px 20px;
    font-size: 17px;
    cursor: pointer;
    font-weight: bold;
    background: transparent;
    position: relative;
    transition: all 1s;
    overflow: hidden;
  }

  .btn:hover {
    color: white;
  }

  .btn::before {
    content: '';
    position: absolute;
    height: 100%;
    width: 0%;
    top: 0;
    left: -40px;
    transform: skewX(45deg);
    background-color: purple;
    z-index: -1;
    transition: all 1s;
  }

  .btn:hover::before {
    width: 160%;
  }
</style>

<script setup lang="ts">
defineProps<{
  showPremarket?: boolean
}>()
</script>

<template>
  <section class="trade-order-tabs">
    <div class="trade-order-tabs__nav">
      <button class="active" type="button">当前委托</button>
      <button type="button">历史订单</button>
      <button v-if="showPremarket" type="button">盘前订单</button>
    </div>
    <p><span>登录</span>或<span>注册</span>查看数据</p>
  </section>
</template>

<style scoped>
.trade-order-tabs__nav {
  display: flex;
  gap: 28px;
  padding: 22px 22px 0;
}

button {
  position: relative;
  padding: 0 0 14px;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 18px;
  font-weight: 500;
}

button.active {
  color: #fff;
}

button.active::after {
  position: absolute;
  right: 6px;
  bottom: 0;
  left: 6px;
  height: 3px;
  border-radius: 999px;
  background: #02b904;
  content: '';
}

p {
  display: grid;
  min-height: 130px;
  margin: 0;
  place-items: center;
  color: #8f929d;
  font-size: 16px;
}

span {
  color: #02b904;
}
</style>

<script setup lang='ts'>
defineProps<{
  showPremarket?: boolean
}>()
</script>

<template>
  <section class="trade-orders-panel">
    <div class="trade-orders-panel__nav">
      <button class="active" type="button">当前委托</button>
      <button type="button">历史订单</button>
      <button v-if="showPremarket" type="button">盘前订单</button>
    </div>
    <p><span>登录</span>或<span>注册</span>查看数据</p>
  </section>
</template>

<style scoped>
.trade-orders-panel__nav {
  display: flex;
  gap: 22px;
  padding: 20px 0 0;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

button {
  position: relative;
  padding: 0 0 14px;
  border: 0;
  background: transparent;
  color: #8f929d;
  font: inherit;
  font-size: 17px;
  font-weight: 700;
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
  min-height: 88px;
  margin: 0;
  place-items: center;
  color: #8f929d;
  font-size: 14px;
}

span {
  color: #02b904;
}

@media (max-width: 390px) {
  .trade-orders-panel__nav {
    gap: 18px;
  }

  button {
    font-size: 15px;
  }
}
</style>

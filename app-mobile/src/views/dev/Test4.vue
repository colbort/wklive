<template>
  <CommonPage>
    <!-- 只有自定义 tabbar -->
    <template #tabbar>
      <div class="custom-tabbar">
        <button type="button" class="tabbar-left" @click="goBack">
          取消
        </button>

        <div class="tabbar-title">
          只有自定义标题栏
        </div>

        <button type="button" class="tabbar-right" @click="onMore">
          更多
        </button>
      </div>
    </template>

    <template #default>
      <div class="page-content">
        <div class="section-title">
          内容列表
        </div>

        <div v-for="i in 50" :key="i" class="list-item">
          <span>content 内容 {{ i }}</span>
          <span class="arrow">›</span>
        </div>
      </div>
    </template>
  </CommonPage>
</template>

<script setup>
import CommonPage from '@/components/common/CommonPage.vue'

function goBack() {
  history.back()
}

function onMore() {
  console.log('更多')
}
</script>

<style scoped>
.custom-tabbar {
  width: 100%;
  height: 58px;
  background: #0b0d16;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-sizing: border-box;
}

.tabbar-left,
.tabbar-right {
  min-width: 56px;
  height: 58px;
  border: none;
  outline: none;
  padding: 0;
  margin: 0;
  background: transparent;
  color: #21ff00;
  font-size: 15px;
  font-weight: 800;
  cursor: pointer;
}

.tabbar-left {
  text-align: left;
}

.tabbar-right {
  text-align: right;
}

.tabbar-title {
  flex: 1;
  text-align: center;
  color: #ffffff;
  font-size: 22px;
  font-weight: 800;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.page-content {
  padding: 18px 24px 110px;
  box-sizing: border-box;
}

.section-title {
  margin-bottom: 16px;
  font-size: 22px;
  font-weight: 800;
  color: #ffffff;
}

.list-item {
  height: 66px;
  margin-bottom: 16px;
  padding: 0 20px 0 24px;
  border-radius: 22px;
  background: #1b1e29;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
}

.arrow {
  font-size: 40px;
  line-height: 1;
  color: #9ca1ad;
  font-weight: 300;
}
</style>

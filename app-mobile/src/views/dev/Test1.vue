<template>
  <CommonPage
    v-model="activeMenu"
    title="页面标题"
    :show-back="true"
    right-text="保存"
    :menus="menus"
    @back="goBack"
    @right-click="onSave"
  >
    <!-- 自定义区域：会跟 tabbar 一起向上滚动 -->
    <template #custom>
      <div class="banner">
        <div class="banner-title">
          自定义区域
        </div>
        <div class="banner-desc">
          这里可以放用户信息、统计卡片、轮播图、筛选条件等内容
        </div>
      </div>
    </template>

    <!-- content：根据 activeKey 展示不同内容 -->
    <template #default="{ activeKey }">
      <div v-if="activeKey === 'one'" class="card">
        <div class="card-title">
          菜单一内容
        </div>

        <div v-for="i in 30" :key="i" class="list-item">
          <span>菜单一内容 {{ i }}</span>
          <span class="arrow">›</span>
        </div>
      </div>

      <div v-else-if="activeKey === 'two'" class="card">
        <div class="card-title">
          菜单二内容
        </div>

        <div v-for="i in 30" :key="i" class="list-item">
          <span>菜单二内容 {{ i }}</span>
          <span class="arrow">›</span>
        </div>
      </div>

      <div v-else class="card">
        <div class="card-title">
          菜单三内容
        </div>

        <div v-for="i in 30" :key="i" class="list-item">
          <span>菜单三内容 {{ i }}</span>
          <span class="arrow">›</span>
        </div>
      </div>
    </template>
  </CommonPage>
</template>

<script setup>
import { ref } from 'vue'
import CommonPage from '@/components/common/CommonPage.vue'

const activeMenu = ref('one')

const menus = [
  { label: '菜单一', value: 'one' },
  { label: '菜单二', value: 'two' },
  { label: '菜单三', value: 'three' },
]

function goBack() {
  history.back()
}

function onSave() {
  console.log('点击右侧按钮')
}
</script>

<style scoped>
.banner {
  margin: 18px 24px 18px;
  height: 180px;
  border-radius: 24px;
  background: #1b1e29;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}

.banner-title {
  font-size: 24px;
  font-weight: 800;
  color: #ffffff;
}

.banner-desc {
  margin-top: 12px;
  padding: 0 28px;
  font-size: 14px;
  line-height: 22px;
  color: #9ca1ad;
  text-align: center;
}

.card {
  padding: 0 24px 110px;
  box-sizing: border-box;
}

.card-title {
  margin-bottom: 16px;
  font-size: 20px;
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

<template>
  <CommonPage v-model="activeMenu" :menus="menus" @back="goBack">
    <!-- 自定义 tabbar -->
    <template #tabbar="{ onBack }">
      <div class="custom-tabbar">
        <button type="button" class="tabbar-left" @click="onBack">
          返回
        </button>

        <div class="tabbar-title">
          自定义Tabbar
        </div>

        <button type="button" class="tabbar-right" @click="onSave">
          保存
        </button>
      </div>
    </template>

    <!-- 内容区域 -->
    <template #default="{ activeKey }">
      <div class="page-content">
        <template v-if="activeKey === 'one'">
          <div class="section-title">
            菜单一内容
          </div>

          <div v-for="i in 30" :key="i" class="list-item">
            <span>菜单一内容 {{ i }}</span>
            <span class="arrow">›</span>
          </div>
        </template>

        <template v-else-if="activeKey === 'two'">
          <div class="section-title">
            菜单二内容
          </div>

          <div v-for="i in 30" :key="i" class="list-item">
            <span>菜单二内容 {{ i }}</span>
            <span class="arrow">›</span>
          </div>
        </template>

        <template v-else>
          <div class="section-title">
            菜单三内容
          </div>

          <div v-for="i in 30" :key="i" class="list-item">
            <span>菜单三内容 {{ i }}</span>
            <span class="arrow">›</span>
          </div>
        </template>
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
  console.log('保存')
}
</script>

<style scoped>
.custom-tabbar {
  width: 100%;
  height: 58px;
  background: var(--page-bg);
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
  color: var(--accent);
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
  color: var(--text);
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
  color: var(--text);
}

.list-item {
  height: 66px;
  margin-bottom: 16px;
  padding: 0 20px 0 24px;
  border-radius: 22px;
  background: var(--panel-bg);
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  font-size: 18px;
  font-weight: 700;
  color: var(--text);
}

.arrow {
  font-size: 40px;
  line-height: 1;
  color: var(--muted);
  font-weight: 300;
}
</style>

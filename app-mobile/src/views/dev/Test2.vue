<template>
  <CommonPage
    title="只有标题栏"
    :show-back="true"
    right-text="更多"
    @back="goBack"
    @right-click="onMore"
  >
    <template #default>
      <div class="page-content">
        <div class="card">
          <div class="card-title">
            内容列表
          </div>

          <div v-for="i in 50" :key="i" class="list-item">
            <span>content 内容 {{ i }}</span>
            <span class="arrow">›</span>
          </div>
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
.page-content {
  padding: 18px 24px 110px;
  box-sizing: border-box;
}

.card {
  width: 100%;
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

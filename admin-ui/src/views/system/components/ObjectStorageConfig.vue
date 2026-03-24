<template>
  <div class="object-storage-config">
    <el-tabs v-model="activeTab" @tab-click="handleTabClick" class="config-tabs">
      <el-tab-pane label="Aliyun OSS" name="aliyun">
        <el-card shadow="never" class="config-card">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.endpoint')">
                <el-input
                  v-model="form.aliyun_oss.endpoint"
                  :placeholder="t('system.endpointPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.accessKeyId')">
                <el-input
                  v-model="form.aliyun_oss.access_key_id"
                  :placeholder="t('system.accessKeyIdPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.accessKeySecret')">
                <el-input
                  v-model="form.aliyun_oss.access_key_secret"
                  :placeholder="t('system.accessKeySecretPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.bucketName')">
                <el-input
                  v-model="form.aliyun_oss.bucket_name"
                  :placeholder="t('system.bucketNamePlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.bucketUrl')">
                <el-input
                  v-model="form.aliyun_oss.bucket_url"
                  :placeholder="t('system.bucketUrlPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="Tencent COS" name="tencent">
        <el-card shadow="never" class="config-card">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.region')">
                <el-input
                  v-model="form.tencent_cos.region"
                  :placeholder="t('system.regionPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.secretId')">
                <el-input
                  v-model="form.tencent_cos.secret_id"
                  :placeholder="t('system.secretIdPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.secretKey')">
                <el-input
                  v-model="form.tencent_cos.secret_key"
                  :placeholder="t('system.secretKeyPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.bucketName')">
                <el-input
                  v-model="form.tencent_cos.bucket_name"
                  :placeholder="t('system.bucketNamePlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.bucketUrl')">
                <el-input
                  v-model="form.tencent_cos.bucket_url"
                  :placeholder="t('system.bucketUrlPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>
      <el-tab-pane label="MinIO" name="minio">
        <el-card shadow="never" class="config-card">
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.endpoint')">
                <el-input
                  v-model="form.minio.endpoint"
                  :placeholder="t('system.endpointPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.accessKeyId')">
                <el-input
                  v-model="form.minio.access_key_id"
                  :placeholder="t('system.accessKeyIdPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.accessKeySecret')">
                <el-input
                  v-model="form.minio.access_key_secret"
                  :placeholder="t('system.accessKeySecretPlaceholder')"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item :label="t('system.bucketName')">
                <el-input
                  v-model="form.minio.bucket_name"
                  :placeholder="t('system.bucketNamePlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item :label="t('system.bucketUrl')">
                <el-input
                  v-model="form.minio.bucket_url"
                  :placeholder="t('system.bucketUrlPlaceholder')"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </el-card>
      </el-tab-pane>
    </el-tabs>
    <el-form-item :label="t('system.ossType')" prop="oss_type">
      <el-select
        v-model="form.oss_type"
        :placeholder="t('system.ossTypePlaceholder')"
        :filterable="false"
      >
        <el-option label="Aliyun OSS" :value="1" />
        <el-option label="Tencent COS" :value="2" />
        <el-option label="MinIO" :value="3" />
      </el-select>
    </el-form-item>
    <el-form-item :label="t('system.ossDomain')" prop="oss_domain">
      <el-input v-model="form.oss_domain" :placeholder="t('common.pleaseEnter')" />
    </el-form-item>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ObjectStorageConfig } from '@/services/system/ConfigService'

const { t } = useI18n()

interface Props {
  modelValue: ObjectStorageConfig
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: ObjectStorageConfig]
}>()

const activeTab = ref('aliyun')

const form = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

function handleTabClick(_tab: any) {
  // 仅切换视图选项卡，不修改 oss_type
}
</script>

<style scoped>
.object-storage-config {
  width: 100%;
}

.config-card {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 0;
  margin-top: 2px;
  margin-bottom: 10px;
}

.config-tabs {
  --el-tabs-header-height: 36px;
}

.config-tabs .el-tabs__header {
  margin: 0 0 10px 0;
}

.config-tabs .el-tabs__nav-wrap::after {
  display: none;
}

.config-tabs .el-tabs__item {
  border-radius: 4px 4px 0 0;
  margin-right: 2px;
  padding: 6px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.config-tabs .el-tabs__item:hover {
  color: #409eff;
  background-color: #ecf5ff;
}

.config-tabs .el-tabs__item.is-active {
  color: #409eff;
  background-color: #ecf5ff;
  border-bottom: 2px solid #409eff;
}

.config-tabs .el-tabs__content {
  padding: 0;
}
</style>

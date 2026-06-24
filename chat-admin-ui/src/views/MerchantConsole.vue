<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import MerchantAgentTable from "@/components/merchant/MerchantAgentTable.vue";
import MerchantCategoryTable from "@/components/merchant/MerchantCategoryTable.vue";
import MerchantGroupTable from "@/components/merchant/MerchantGroupTable.vue";

type TabName = "agents" | "categories" | "groups";

const route = useRoute();
const activeTab = computed<TabName>(
  () => String(route.meta.activeTab || "agents") as TabName,
);
</script>

<template>
  <section class="console-page">
    <div class="table-wrap">
      <MerchantAgentTable v-if="activeTab === 'agents'" />
      <MerchantCategoryTable v-else-if="activeTab === 'categories'" />
      <MerchantGroupTable v-else />
    </div>
  </section>
</template>

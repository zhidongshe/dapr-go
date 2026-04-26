<template>
  <div class="inventory-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>库存概览</span>
          <el-button type="primary" @click="fetchData">刷新</el-button>
        </div>
      </template>

      <!-- 统计卡片 -->
      <el-row :gutter="20" class="stats-row">
        <el-col :span="8">
          <div class="stat-box">
            <div class="stat-number">{{ inventoryList.length }}</div>
            <div class="stat-label">总商品数</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box warning">
            <div class="stat-number">{{ warningList.length }}</div>
            <div class="stat-label">库存预警 (&lt;20)</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box danger">
            <div class="stat-number">{{ outOfStockCount }}</div>
            <div class="stat-label">缺货商品</div>
          </div>
        </el-col>
      </el-row>

      <!-- 预警列表 -->
      <h4 v-if="warningList.length" class="section-title">库存预警</h4>
      <el-table v-if="warningList.length" :data="warningList" stripe>
        <el-table-column prop="product_id" label="商品ID" width="100" />
        <el-table-column prop="product_name" label="商品名称" />
        <el-table-column prop="available_stock" label="可用库存" width="120">
          <template #default="{ row }">
            <span style="color: #F56C6C; font-weight: bold;">{{ row.available_stock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="reserved_stock" label="已预留" width="120" />
      </el-table>

      <!-- 库存列表 -->
      <h4 class="section-title" style="margin-top: 30px;">库存列表</h4>
      <el-table :data="inventoryList" v-loading="loading" stripe>
        <el-table-column prop="product_id" label="商品ID" width="100" />
        <el-table-column prop="product_name" label="商品名称" />
        <el-table-column label="总库存" width="100">
          <template #default="{ row }">
            {{ row.available_stock + row.reserved_stock }}
          </template>
        </el-table-column>
        <el-table-column prop="reserved_stock" label="已预留" width="100" />
        <el-table-column prop="available_stock" label="可用库存" width="120">
          <template #default="{ row }">
            <span :style="{ color: row.available_stock < 20 ? '#F56C6C' : '#67C23A' }">
              {{ row.available_stock }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import request from '@/api/request'

interface InventoryItem {
  product_id: number
  product_name: string
  available_stock: number
  reserved_stock: number
  version: number
  updated_at: string
}

const loading = ref(false)
const inventoryList = ref<InventoryItem[]>([])

const warningList = computed(() =>
  inventoryList.value.filter(item => item.available_stock < 20 && item.available_stock > 0)
)

const outOfStockCount = computed(() =>
  inventoryList.value.filter(item => item.available_stock <= 0).length
)

const fetchData = async () => {
  loading.value = true
  try {
    const res = await request.get('/inventory')
    inventoryList.value = res.data.data || []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.inventory-page {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stats-row {
  margin-bottom: 30px;
}

.stat-box {
  background: #f5f7fa;
  padding: 20px;
  text-align: center;
  border-radius: 8px;
}

.stat-box.warning {
  background: #fdf6ec;
}

.stat-box.danger {
  background: #fef0f0;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  color: #409EFF;
  margin-bottom: 10px;
}

.stat-box.warning .stat-number {
  color: #E6A23C;
}

.stat-box.danger .stat-number {
  color: #F56C6C;
}

.stat-label {
  color: #606266;
  font-size: 14px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  margin-bottom: 15px;
  padding-left: 10px;
  border-left: 4px solid #409EFF;
}
</style>

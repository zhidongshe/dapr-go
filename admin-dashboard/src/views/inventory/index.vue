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
            <div class="stat-number">500</div>
            <div class="stat-label">总商品数</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box warning">
            <div class="stat-number">10</div>
            <div class="stat-label">库存预警</div>
          </div>
        </el-col>
        <el-col :span="8">
          <div class="stat-box danger">
            <div class="stat-number">2</div>
            <div class="stat-label">缺货商品</div>
          </div>
        </el-col>
      </el-row>

      <!-- 预警列表 -->
      <h4 class="section-title">库存预警</h4>
      <el-table :data="warningList" stripe>
        <el-table-column prop="productId" label="商品ID" width="100" />
        <el-table-column prop="productName" label="商品名称" />
        <el-table-column prop="currentStock" label="当前库存" width="120">
          <template #default="{ row }">
            <span style="color: #F56C6C; font-weight: bold;">{{ row.currentStock }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="warningThreshold" label="预警阈值" width="120" />
      </el-table>

      <!-- 库存列表 -->
      <h4 class="section-title" style="margin-top: 30px;">库存列表</h4>
      <el-table :data="inventoryList" stripe>
        <el-table-column prop="productId" label="商品ID" width="100" />
        <el-table-column prop="productName" label="商品名称" />
        <el-table-column prop="stock" label="总库存" width="100" />
        <el-table-column prop="reserved" label="已预留" width="100" />
        <el-table-column prop="available" label="可用库存" width="120">
          <template #default="{ row }">
            <span :style="{ color: row.available < row.warningThreshold ? '#F56C6C' : '#67C23A' }">
              {{ row.available }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="warningThreshold" label="预警阈值" width="120" />
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const page = ref(1)
const pageSize = ref(10)
const total = ref(100)

// 模拟数据
const warningList = ref([
  { productId: 101, productName: 'iPhone 15 手机壳', currentStock: 5, warningThreshold: 10 },
  { productId: 102, productName: 'AirPods Pro 保护套', currentStock: 3, warningThreshold: 10 },
  { productId: 103, productName: 'Type-C 数据线', currentStock: 8, warningThreshold: 20 }
])

const inventoryList = ref([
  { productId: 101, productName: 'iPhone 15 手机壳', stock: 50, reserved: 10, available: 40, warningThreshold: 10 },
  { productId: 102, productName: 'AirPods Pro 保护套', stock: 30, reserved: 5, available: 25, warningThreshold: 10 },
  { productId: 103, productName: 'Type-C 数据线', stock: 100, reserved: 20, available: 80, warningThreshold: 20 }
])

const fetchData = () => {
  // TODO: 调用 API 获取数据
}

const handleCurrentChange = (val: number) => {
  page.value = val
  fetchData()
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

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

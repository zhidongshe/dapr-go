<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon" style="background: #409EFF;">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.orders.total || 0 }}</div>
            <div class="stat-label">总订单数</div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon" style="background: #67C23A;">
            <el-icon><Money /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">¥{{ formatMoney(stats?.payments.todayAmount || 0) }}</div>
            <div class="stat-label">今日销售额</div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon" style="background: #E6A23C;">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.orders.pending || 0 }}</div>
            <div class="stat-label">待处理订单</div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon" style="background: #F56C6C;">
            <el-icon><Warning /></el-icon>
          </div>
          <div class="stat-info">
            <div class="stat-value">{{ stats?.inventory.warningCount || 0 }}</div>
            <div class="stat-label">库存预警</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="chart-row">
      <el-col :span="12">
        <el-card>
          <template #header>订单状态分布</template>
          <div class="chart-placeholder">
            <el-row :gutter="10">
              <el-col :span="12">
                <div class="status-item">
                  <span class="status-dot" style="background: #E6A23C;"></span>
                  <span>待支付: {{ stats?.orders.pending || 0 }}</span>
                </div>
                <div class="status-item">
                  <span class="status-dot" style="background: #409EFF;"></span>
                  <span>已支付: {{ stats?.orders.paid || 0 }}</span>
                </div>
                <div class="status-item">
                  <span class="status-dot" style="background: #67C23A;"></span>
                  <span>处理中: {{ stats?.orders.processing || 0 }}</span>
                </div>
              </el-col>
              <el-col :span="12">
                <div class="status-item">
                  <span class="status-dot" style="background: #909399;"></span>
                  <span>已发货: {{ stats?.orders.shipped || 0 }}</span>
                </div>
                <div class="status-item">
                  <span class="status-dot" style="background: #67C23A;"></span>
                  <span>已完成: {{ stats?.orders.completed || 0 }}</span>
                </div>
                <div class="status-item">
                  <span class="status-dot" style="background: #F56C6C;"></span>
                  <span>已取消: {{ stats?.orders.cancelled || 0 }}</span>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>销售趋势（本周）</template>
          <div class="chart-placeholder">
            <div class="trend-item">
              <span>今日订单: {{ stats?.payments.todayCount || 0 }} 笔</span>
            </div>
            <div class="trend-item">
              <span>本周销售额: ¥{{ formatMoney(stats?.payments.weekAmount || 0) }}</span>
            </div>
            <div class="trend-item">
              <span>本月销售额: ¥{{ formatMoney(stats?.payments.monthAmount || 0) }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboardStats } from '@/api/dashboard'
import type { DashboardStats } from '@/types/api'

const stats = ref<DashboardStats | null>(null)
const loading = ref(false)

const formatMoney = (amount: number) => {
  return (amount / 100).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const fetchStats = async () => {
  loading.value = true
  try {
    const response = await getDashboardStats()
    stats.value = response.data.data
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStats()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 10px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30px;
  color: #fff;
  margin-right: 15px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.chart-row {
  margin-top: 20px;
}

.chart-placeholder {
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.status-item {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 10px;
}

.trend-item {
  padding: 15px 0;
  border-bottom: 1px solid #ebeef5;
}

.trend-item:last-child {
  border-bottom: none;
}
</style>

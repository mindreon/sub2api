export default {
  common: {
    apply: '应用',
    clear: '清空',
    creating: '创建中...',
    required: '必填',
    sending: '发送中...',
    tryAgain: '重试',
  },
  admin: {
    accounts: {
      fromModel: '源模型',
      toModel: '目标模型',
      messages: {
        accountCreated: '账号创建成功',
      },
      oauth: {
        openai: {
          accessTokenAuth: 'Access Token 授权',
          mobileRefreshTokenAuth: '移动端 Refresh Token 授权',
        },
      },
    },
    channels: {
      noGroupsSelected: '请至少选择一个分组',
      emptyModelsInPricing: '请至少填写一个模型',
    },
    distribution: {
      tabs: {
        walletTransactions: '钱包流水',
      },
    },
    ops: {
      runtime: {
        metricThresholds: '指标阈值配置',
        metricThresholdsHint: '配置各项指标的告警阈值，超出阈值时将以红色显示',
        slaMinPercent: 'SLA 最低百分比',
        slaMinPercentHint: 'SLA 低于此值时显示为红色（默认：99.5%）',
        ttftP99MaxMs: 'TTFT P99 最大值（毫秒）',
        ttftP99MaxMsHint: 'TTFT P99 高于此值时显示为红色（默认：500ms）',
        requestErrorRateMaxPercent: '请求错误率最大值（%）',
        requestErrorRateMaxPercentHint: '请求错误率高于此值时显示为红色（默认：5%）',
        upstreamErrorRateMaxPercent: '上游错误率最大值（%）',
        upstreamErrorRateMaxPercentHint: '上游错误率高于此值时显示为红色（默认：5%）',
      },
    },
    settings: {
      openaiFastPolicy: {
        userIds: '用户 ID 范围',
        userIdsHint: '指定规则仅对这些用户 ID 生效；留空表示所有用户。',
        userIdPlaceholder: '输入用户 ID',
        removeUserId: '移除用户 ID',
        addUserId: '添加用户 ID',
      },
    },
    users: {
      passwordCopied: '密码已复制',
    },
  },
  distribution: {
    columns: {
      alertType: '预警类型',
      severity: '严重级别',
      summary: '摘要',
      triggeredAt: '触发时间',
      resolvedAt: '恢复时间',
      lastObservedAt: '最后观察时间',
    },
  },
} as const

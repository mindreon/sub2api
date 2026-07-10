export default {
  common: {
    apply: 'Apply',
    clear: 'Clear',
    creating: 'Creating...',
    required: 'Required',
    sending: 'Sending...',
    tryAgain: 'Try again',
  },
  admin: {
    accounts: {
      fromModel: 'From model',
      toModel: 'To model',
      messages: {
        accountCreated: 'Account created',
      },
      oauth: {
        openai: {
          accessTokenAuth: 'Access token auth',
          mobileRefreshTokenAuth: 'Mobile refresh token auth',
        },
      },
    },
    channels: {
      noGroupsSelected: 'Select at least one group',
      emptyModelsInPricing: 'Enter at least one model',
    },
    distribution: {
      tabs: {
        walletTransactions: 'Wallet Transactions',
      },
    },
    ops: {
      runtime: {
        metricThresholds: 'Metric Thresholds',
        metricThresholdsHint: 'Configure alert thresholds for metrics. Values exceeding thresholds are shown in red.',
        slaMinPercent: 'SLA Minimum Percentage',
        slaMinPercentHint: 'SLA below this value is shown in red (default: 99.5%).',
        ttftP99MaxMs: 'TTFT P99 Maximum (ms)',
        ttftP99MaxMsHint: 'TTFT P99 above this value is shown in red (default: 500ms).',
        requestErrorRateMaxPercent: 'Request Error Rate Maximum (%)',
        requestErrorRateMaxPercentHint: 'Request error rate above this value is shown in red (default: 5%).',
        upstreamErrorRateMaxPercent: 'Upstream Error Rate Maximum (%)',
        upstreamErrorRateMaxPercentHint: 'Upstream error rate above this value is shown in red (default: 5%).',
      },
    },
    settings: {
      openaiFastPolicy: {
        userIds: 'User ID scope',
        userIdsHint: 'Apply this rule only to these user IDs. Leave empty for all users.',
        userIdPlaceholder: 'Enter user ID',
        removeUserId: 'Remove user ID',
        addUserId: 'Add user ID',
      },
    },
    users: {
      passwordCopied: 'Password copied',
    },
  },
  distribution: {
    columns: {
      alertType: 'Alert Type',
      severity: 'Severity',
      summary: 'Summary',
      triggeredAt: 'Triggered At',
      resolvedAt: 'Resolved At',
      lastObservedAt: 'Last Observed At',
    },
  },
} as const

<template>
  <AppLayout>
    <div class="yk">
      <div data-test="api-address" class="yk-address">
        <span class="yk-address-label">{{ t('yc.connect.baseUrl') }}</span>
        <code class="yk-mono yk-address-url">{{ apiAddress }}</code>
        <button type="button" class="yk-btn yk-btn-ghost" @click="clipboardCopy(apiAddress, t('keys.copied'))">
          <Icon name="copy" size="sm" />
          {{ t('yc.connect.copyShort') }}
        </button>
        <p class="yk-address-hint">
          {{ t('keys.addressHint') }}
          <router-link to="/help/guide" class="yk-link">{{ t('yc.connect.toolGuide') }}</router-link>
        </p>
      </div>

      <div class="yk-tools">
        <SearchInput
          v-model="filterSearch"
          :placeholder="t('keys.searchPlaceholder')"
          class="w-full sm:w-60"
          @search="onFilterChange"
        />
        <Select
          :model-value="filterGroupId"
          class="w-36"
          :options="groupFilterOptions"
          @update:model-value="onGroupFilterChange"
        />
        <Select
          :model-value="filterStatus"
          class="w-36"
          :options="statusFilterOptions"
          @update:model-value="onStatusFilterChange"
        />
        <Select
          data-test="key-sort"
          :model-value="sortValue"
          class="w-36"
          :options="sortOptions"
          @update:model-value="onSortChange"
        />
        <EndpointPopover
          v-if="publicSettings?.api_base_url || (publicSettings?.custom_endpoints?.length ?? 0) > 0"
          :api-base-url="publicSettings?.api_base_url || ''"
          :custom-endpoints="publicSettings?.custom_endpoints || []"
        />
        <span class="yk-grow"></span>
        <button
          type="button"
          class="yk-btn yk-btn-ghost yk-btn-icon"
          :disabled="loading"
          :title="t('common.refresh')"
          :aria-label="t('common.refresh')"
          @click="loadApiKeys"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button type="button" class="yk-btn yk-btn-lime" data-tour="keys-create-btn" @click="showCreateModal = true">
          <Icon name="plus" size="md" />
          {{ t('keys.createKey') }}
        </button>
      </div>

      <div v-if="selectedIds.length" class="yk-bulk">
        <span>{{ t('keys.bulkEdit.selectedCount', { count: selectedIds.length }) }}</span>
        <button
          type="button"
          class="yk-btn yk-btn-ink"
          :disabled="loading"
          data-test="bulk-edit-keys"
          @click="showBulkEditModal = true"
        >
          {{ t('keys.bulkEdit.title') }}
        </button>
        <button type="button" class="yk-btn yk-btn-ghost" data-test="select-all-keys" @click="selectAllVisible">
          {{ t('keys.selectAllOnPage') }}
        </button>
        <button type="button" class="yk-btn yk-btn-ghost" @click="selectedIds = []">
          {{ t('keys.bulkEdit.clearSelection') }}
        </button>
      </div>

      <div class="yk-list" data-test="key-list">
        <template v-if="loading && apiKeys.length === 0">
          <div v-for="i in 3" :key="i" class="yk-card yk-skeleton" aria-hidden="true"><i></i><i></i><i></i></div>
        </template>

        <div v-else-if="apiKeys.length === 0" class="yk-card yk-empty">
          <EmptyState
            :title="t('keys.noKeysYet')"
            :description="t('keys.createFirstKey')"
            :action-text="t('keys.createKey')"
            @action="showCreateModal = true"
          />
        </div>

        <template v-else>
        <article
          v-for="row in apiKeys"
          :key="row.id"
          data-test="key-card"
          class="yk-card"
          :class="{ 'yk-card-off': row.status !== 'active', 'yk-card-picked': selectedIds.includes(row.id) }"
        >
          <header class="yk-head">
            <input
              type="checkbox"
              class="yk-check"
              data-test="key-select"
              :checked="selectedIds.includes(row.id)"
              :aria-label="t('keys.bulkEdit.selectKey', { name: row.name })"
              @change="toggleSelected(row.id)"
            />
            <h3 class="yk-name">{{ row.name }}</h3>
            <Icon
              v-if="(row.ip_whitelist?.length ?? 0) > 0 || (row.ip_blacklist?.length ?? 0) > 0"
              name="shield"
              size="sm"
              class="yk-shield"
              :title="t('keys.ipRestrictionEnabled')"
            />
            <span class="yk-status" :class="'yk-status-' + row.status">{{ t('keys.status.' + row.status) }}</span>
            <span class="yk-grow"></span>
            <button
              v-if="!publicSettings?.hide_ccs_import_button"
              type="button"
              data-test="key-ccs"
              class="yk-btn yk-btn-ghost"
              @click="importToCcswitch(row)"
            >
              <Icon name="upload" size="sm" />
              {{ t('keys.importToCcSwitch') }}
            </button>
            <button
              type="button"
              data-test="key-more"
              class="action-menu-trigger yk-btn yk-btn-ghost yk-btn-icon"
              :class="{ 'yk-btn-on': activeMenuId === row.id }"
              :aria-label="t('common.more')"
              :title="t('common.more')"
              aria-haspopup="menu"
              :aria-expanded="activeMenuId === row.id"
              @click="openActionMenu(row, $event)"
            >
              <Icon name="more" size="sm" />
            </button>
          </header>

          <div class="yk-key">
            <code class="yk-mono yk-key-value">{{ maskApiKey(row.key) }}</code>
            <button
              type="button"
              data-test="key-copy"
              class="yk-btn yk-btn-lime"
              @click="copyToClipboard(row.key, row.id)"
            >
              <Icon :name="copiedKeyId === row.id ? 'check' : 'copy'" size="sm" :stroke-width="2" />
              {{ copiedKeyId === row.id ? t('keys.copied') : t('keys.copyKey') }}
            </button>
          </div>

          <dl class="yk-facts">
            <div class="yk-fact group/dropdown">
              <dt>{{ t('yc.prices.line') }}</dt>
              <dd>
                <button
                  :ref="(el) => setGroupButtonRef(row.id, el)"
                  type="button"
                  class="yk-line"
                  :title="t('keys.clickToChangeGroup')"
                  @click="openGroupSelector(row)"
                >
                  <template v-if="row.group">
                    <span class="yk-line-name">{{ row.group.name }}</span>
                    <span v-if="row.group.app_only" class="yk-tag">{{ t('yc.prices.appOnly') }}</span>
                    <span v-if="linePriceTag(row)" data-test="key-line-price" class="yk-tag">{{ linePriceTag(row) }}</span>
                  </template>
                  <span v-else class="yk-dim">{{ t('keys.noGroup') }}</span>
                  <span class="yk-line-change">{{ t('keys.changeLine') }}</span>
                </button>
              </dd>
            </div>
            <div class="yk-fact" data-test="usage">
              <dt>{{ t('keys.today') }}</dt>
              <dd>
                <span v-if="usageLoading && !usageStats[row.id]" data-test="usage-pending" class="yk-pending"></span>
                <span v-else class="yk-num">{{ sign }}{{ (usageStats[row.id]?.today_actual_cost ?? 0).toFixed(4) }}</span>
              </dd>
              <dt>{{ t('keys.total') }}</dt>
              <dd>
                <span v-if="usageLoading && !usageStats[row.id]" class="yk-pending"></span>
                <span v-else class="yk-num">{{ sign }}{{ (usageStats[row.id]?.total_actual_cost ?? 0).toFixed(4) }}</span>
              </dd>
            </div>
          </dl>

          <div v-if="row.quota > 0" class="yk-quota text-sm">
            <div class="flex items-center gap-1.5">
              <span class="text-gray-500 dark:text-gray-400">{{ t('keys.quota') }}:</span>
              <span :class="[
                'font-medium',
                row.quota_used >= row.quota ? 'text-red-500' :
                row.quota_used >= row.quota * 0.8 ? 'text-yellow-500' :
                'text-gray-900 dark:text-white'
              ]">
                {{ sign }}{{ row.quota_used?.toFixed(2) || '0.00' }} / {{ sign }}{{ row.quota?.toFixed(2) }}
              </span>
            </div>
            <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
              <div
                :class="[
                  'h-full rounded-full transition-all',
                  row.quota_used >= row.quota ? 'bg-red-500' :
                  row.quota_used >= row.quota * 0.8 ? 'bg-yellow-500' :
                  'bg-primary-500'
                ]"
                :style="{ width: Math.min((row.quota_used / row.quota) * 100, 100) + '%' }"
              />
            </div>
          </div>

          <details class="yk-details" data-test="key-details">
            <summary>{{ t('keys.details') }}</summary>
            <dl class="yk-detail-grid">
              <div v-if="row.rate_limit_5h > 0 || row.rate_limit_1d > 0 || row.rate_limit_7d > 0" class="yk-detail-wide">
                <dt>{{ t('keys.rateLimitColumn') }}</dt>
                <dd>
                  <div class="space-y-1.5 min-w-[140px] max-w-xs">
                    <!-- 5h window -->
                    <div v-if="row.rate_limit_5h > 0">
                      <div class="flex items-center justify-between text-xs">
                        <span class="text-gray-500 dark:text-gray-400">5h</span>
                        <span :class="[
                          'font-medium tabular-nums',
                          row.usage_5h >= row.rate_limit_5h ? 'text-red-500' :
                          row.usage_5h >= row.rate_limit_5h * 0.8 ? 'text-yellow-500' :
                          'text-gray-700 dark:text-gray-300'
                        ]">
                          {{ sign }}{{ row.usage_5h?.toFixed(2) || '0.00' }}/{{ sign }}{{ row.rate_limit_5h?.toFixed(2) }}
                        </span>
                      </div>
                      <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                        <div
                          :class="[
                            'h-full rounded-full transition-all',
                            row.usage_5h >= row.rate_limit_5h ? 'bg-red-500' :
                            row.usage_5h >= row.rate_limit_5h * 0.8 ? 'bg-yellow-500' :
                            'bg-emerald-500'
                          ]"
                          :style="{ width: Math.min((row.usage_5h / row.rate_limit_5h) * 100, 100) + '%' }"
                        />
                      </div>
                      <div v-if="row.reset_5h_at && formatResetTime(row.reset_5h_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                        ⟳ {{ formatResetTime(row.reset_5h_at) }}
                      </div>
                    </div>
                    <!-- 1d window -->
                    <div v-if="row.rate_limit_1d > 0">
                      <div class="flex items-center justify-between text-xs">
                        <span class="text-gray-500 dark:text-gray-400">1d</span>
                        <span :class="[
                          'font-medium tabular-nums',
                          row.usage_1d >= row.rate_limit_1d ? 'text-red-500' :
                          row.usage_1d >= row.rate_limit_1d * 0.8 ? 'text-yellow-500' :
                          'text-gray-700 dark:text-gray-300'
                        ]">
                          {{ sign }}{{ row.usage_1d?.toFixed(2) || '0.00' }}/{{ sign }}{{ row.rate_limit_1d?.toFixed(2) }}
                        </span>
                      </div>
                      <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                        <div
                          :class="[
                            'h-full rounded-full transition-all',
                            row.usage_1d >= row.rate_limit_1d ? 'bg-red-500' :
                            row.usage_1d >= row.rate_limit_1d * 0.8 ? 'bg-yellow-500' :
                            'bg-emerald-500'
                          ]"
                          :style="{ width: Math.min((row.usage_1d / row.rate_limit_1d) * 100, 100) + '%' }"
                        />
                      </div>
                      <div v-if="row.reset_1d_at && formatResetTime(row.reset_1d_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                        ⟳ {{ formatResetTime(row.reset_1d_at) }}
                      </div>
                    </div>
                    <!-- 7d window -->
                    <div v-if="row.rate_limit_7d > 0">
                      <div class="flex items-center justify-between text-xs">
                        <span class="text-gray-500 dark:text-gray-400">7d</span>
                        <span :class="[
                          'font-medium tabular-nums',
                          row.usage_7d >= row.rate_limit_7d ? 'text-red-500' :
                          row.usage_7d >= row.rate_limit_7d * 0.8 ? 'text-yellow-500' :
                          'text-gray-700 dark:text-gray-300'
                        ]">
                          {{ sign }}{{ row.usage_7d?.toFixed(2) || '0.00' }}/{{ sign }}{{ row.rate_limit_7d?.toFixed(2) }}
                        </span>
                      </div>
                      <div class="h-1 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                        <div
                          :class="[
                            'h-full rounded-full transition-all',
                            row.usage_7d >= row.rate_limit_7d ? 'bg-red-500' :
                            row.usage_7d >= row.rate_limit_7d * 0.8 ? 'bg-yellow-500' :
                            'bg-emerald-500'
                          ]"
                          :style="{ width: Math.min((row.usage_7d / row.rate_limit_7d) * 100, 100) + '%' }"
                        />
                      </div>
                      <div v-if="row.reset_7d_at && formatResetTime(row.reset_7d_at)" class="text-[10px] text-gray-400 dark:text-gray-500 tabular-nums">
                        ⟳ {{ formatResetTime(row.reset_7d_at) }}
                      </div>
                    </div>
                    <!-- Reset button -->
                    <button
                      v-if="row.usage_5h > 0 || row.usage_1d > 0 || row.usage_7d > 0"
                      @click.stop="confirmResetRateLimitFromTable(row)"
                      class="mt-0.5 inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                      :title="t('keys.resetRateLimitUsage')"
                    >
                      <Icon name="refresh" size="xs" />
                      {{ t('keys.resetUsage') }}
                    </button>
                  </div>
                </dd>
              </div>
              <div>
                <dt>{{ t('keys.currentConcurrency') }}</dt>
                <dd data-test="current-concurrency" class="yk-num">{{ row.current_concurrency ?? 0 }}</dd>
              </div>
              <div>
                <dt>{{ t('keys.expiresAt') }}</dt>
                <dd :class="{ 'yk-bad': row.expires_at && new Date(row.expires_at) < new Date() }">
                  {{ row.expires_at ? formatDateTime(row.expires_at) : t('keys.noExpiration') }}
                </dd>
              </div>
              <div>
                <dt>{{ t('keys.lastUsedAt') }}</dt>
                <dd>{{ row.last_used_at ? formatDateTime(row.last_used_at) : '-' }}</dd>
              </div>
              <div>
                <dt>{{ t('keys.lastUsedIP') }}</dt>
                <dd data-test="last-used-ip">{{ row.last_used_ip || '-' }}</dd>
              </div>
              <div>
                <dt>{{ t('keys.created') }}</dt>
                <dd>{{ formatDateTime(row.created_at) }}</dd>
              </div>
              <div>
                <dt>{{ t('keys.id') }}</dt>
                <dd data-test="key-id" class="yk-mono">#{{ row.id }}</dd>
              </div>
            </dl>
          </details>
        </article>
        </template>
      </div>

      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showCreateModal || showEditModal"
      :title="showEditModal ? t('keys.editKey') : t('keys.createKey')"
      width="normal"
      @close="closeModals"
    >
      <form id="key-form" @submit.prevent="handleSubmit" class="space-y-5">
        <div>
          <label class="input-label">{{ t('keys.nameLabel') }}</label>
          <input
            v-model="formData.name"
            type="text"
            required
            class="input"
            :placeholder="t('keys.namePlaceholder')"
            data-tour="key-form-name"
          />
        </div>

        <fieldset v-if="!showEditModal" data-tour="key-form-provider">
          <legend class="input-label">{{ t('keys.providerLabel') }}</legend>
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <label
              v-for="provider in createProviderOptions"
              :key="provider.value"
              class="relative min-w-0"
              :class="provider.count === 0 ? 'cursor-not-allowed' : 'cursor-pointer'"
            >
              <input
                type="radio"
                name="key-provider"
                :value="provider.value"
                :checked="createProvider === provider.value"
                :disabled="provider.count === 0"
                class="peer sr-only"
                @change="selectCreateProvider(provider.value)"
              />
              <span
                class="flex h-full flex-col items-center gap-2 rounded-xl border border-gray-200 bg-white px-2 py-3 text-center transition-colors peer-checked:border-primary-500 peer-checked:bg-primary-50/60 peer-checked:ring-1 peer-checked:ring-primary-500 peer-focus-visible:outline peer-focus-visible:outline-2 peer-focus-visible:outline-offset-2 peer-focus-visible:outline-primary-500 peer-disabled:opacity-40 dark:border-dark-600 dark:bg-dark-800 dark:peer-checked:border-primary-500 dark:peer-checked:bg-primary-500/10"
                :class="provider.count > 0 && 'hover:border-primary-300 dark:hover:border-primary-700'"
              >
                <span class="flex h-8 items-center justify-center gap-1.5" aria-hidden="true">
                  <span
                    v-for="platform in KEY_GROUP_PROVIDER_ICONS[provider.value]"
                    :key="platform"
                    class="flex h-8 w-8 items-center justify-center rounded-lg"
                    :class="platformBadgeLightClass(platform)"
                  >
                    <PlatformIcon :platform="platform" size="lg" />
                  </span>
                </span>
                <span class="text-sm font-semibold text-gray-800 dark:text-gray-100">{{ provider.label }}</span>
              </span>
              <span
                v-if="createProvider === provider.value"
                class="absolute right-1.5 top-1.5 flex h-4 w-4 items-center justify-center rounded-full bg-primary-500 text-white"
                aria-hidden="true"
              >
                <Icon name="check" size="xs" :stroke-width="3" />
              </span>
            </label>
          </div>
          <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400" aria-live="polite">
            {{ groups.length === 0 ? t('common.noGroupsAvailable') : t(`keys.providerHints.${createProvider}`) }}
          </p>
        </fieldset>

        <div>
          <label class="input-label" for="key-form-group">{{ t('keys.groupLabel') }}</label>
          <Select
            :key="showEditModal ? 'edit' : createProvider"
            id="key-form-group"
            :aria-label="t('keys.groupLabel')"
            v-model="formData.group_id"
            :options="formGroupOptions"
            :placeholder="t('keys.selectGroup')"
            :empty-text="t('common.noGroupsAvailable')"
            :searchable="true"
            :search-placeholder="t('keys.searchGroup')"
            data-tour="key-form-group"
          >
            <template #selected="{ option }">
              <GroupBadge
                v-if="option"
                :name="(option as unknown as GroupOption).label"
                :platform="(option as unknown as GroupOption).platform"
                :subscription-type="(option as unknown as GroupOption).subscriptionType"
                :rate-multiplier="(option as unknown as GroupOption).rate"
                :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                :peak-start="(option as unknown as GroupOption).peakStart"
                :peak-end="(option as unknown as GroupOption).peakEnd"
                :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
              />
              <span v-else class="text-gray-400">{{ t('keys.selectGroup') }}</span>
            </template>
            <template #option="{ option, selected }">
              <GroupOptionItem
                :name="(option as unknown as GroupOption).label"
                :platform="(option as unknown as GroupOption).platform"
                :subscription-type="(option as unknown as GroupOption).subscriptionType"
                :rate-multiplier="(option as unknown as GroupOption).rate"
                :user-rate-multiplier="(option as unknown as GroupOption).userRate"
                :peak-rate-enabled="(option as unknown as GroupOption).peakRateEnabled"
                :peak-start="(option as unknown as GroupOption).peakStart"
                :peak-end="(option as unknown as GroupOption).peakEnd"
                :peak-rate-multiplier="(option as unknown as GroupOption).peakRateMultiplier"
                :description="(option as unknown as GroupOption).description"
                :selected="selected"
              />
            </template>
          </Select>
        </div>

        <!-- Custom Key Section (only for create) -->
        <div v-if="!showEditModal" class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.customKeyLabel') }}</label>
            <button
              type="button"
              @click="formData.use_custom_key = !formData.use_custom_key"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.use_custom_key ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.use_custom_key ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>
          <div v-if="formData.use_custom_key">
            <input
              v-model="formData.custom_key"
              type="text"
              class="input font-mono"
              :placeholder="t('keys.customKeyPlaceholder')"
              :class="{ 'border-red-500 dark:border-red-500': customKeyError }"
            />
            <p v-if="customKeyError" class="mt-1 text-sm text-red-500">{{ customKeyError }}</p>
            <p v-else class="input-hint">{{ t('keys.customKeyHint') }}</p>
          </div>
        </div>

        <div v-if="showEditModal">
          <label class="input-label">{{ t('keys.statusLabel') }}</label>
          <Select
            v-model="formData.status"
            :options="statusOptions"
            :placeholder="t('keys.selectStatus')"
          />
        </div>

        <!-- IP Restriction Section -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.ipRestriction') }}</label>
            <button
              type="button"
              @click="formData.enable_ip_restriction = !formData.enable_ip_restriction"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_ip_restriction ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_ip_restriction ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_ip_restriction" class="space-y-4 pt-2">
            <div>
              <label class="input-label">{{ t('keys.ipWhitelist') }}</label>
              <textarea
                v-model="formData.ip_whitelist"
                rows="3"
                class="input font-mono text-sm"
                :placeholder="t('keys.ipWhitelistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipWhitelistHint') }}</p>
            </div>

            <div>
              <label class="input-label">{{ t('keys.ipBlacklist') }}</label>
              <textarea
                v-model="formData.ip_blacklist"
                rows="3"
                class="input font-mono text-sm"
                :placeholder="t('keys.ipBlacklistPlaceholder')"
              />
              <p class="input-hint">{{ t('keys.ipBlacklistHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Quota Limit Section -->
        <div class="space-y-3">
          <label class="input-label">{{ t('keys.quotaLimit') }}</label>
          <!-- Switch commented out - always show input, 0 = unlimited
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.quotaLimit') }}</label>
            <button
              type="button"
              @click="formData.enable_quota = !formData.enable_quota"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_quota ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_quota ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>
          -->

          <div class="space-y-4">
            <div>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ sign }}</span>
                <input
                  v-model.number="formData.quota"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="t('keys.quotaAmountPlaceholder')"
                />
              </div>
              <p class="input-hint">{{ t('keys.quotaAmountHint') }}</p>
            </div>

            <!-- Quota used display (only in edit mode) -->
            <div v-if="showEditModal && selectedKey && selectedKey.quota > 0">
              <label class="input-label">{{ t('keys.quotaUsed') }}</label>
              <div class="flex items-center gap-2">
                <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700">
                  <span class="font-medium text-gray-900 dark:text-white">
                    {{ sign }}{{ selectedKey.quota_used?.toFixed(4) || '0.0000' }}
                  </span>
                  <span class="mx-2 text-gray-400">/</span>
                  <span class="text-gray-500 dark:text-gray-400">
                    {{ sign }}{{ selectedKey.quota?.toFixed(2) || '0.00' }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="confirmResetQuota"
                  class="btn btn-secondary text-sm"
                  :title="t('keys.resetQuotaUsed')"
                >
                  {{ t('keys.reset') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Rate Limit Section -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.rateLimitSection') }}</label>
            <button
              type="button"
              @click="formData.enable_rate_limit = !formData.enable_rate_limit"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_rate_limit ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_rate_limit ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_rate_limit" class="space-y-4 pt-2">
            <p class="input-hint -mt-2">{{ t('keys.rateLimitHint') }}</p>
            <!-- 5-Hour Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit5h') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ sign }}</span>
                <input
                  v-model.number="formData.rate_limit_5h"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_5h > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'text-red-500' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ sign }}{{ selectedKey.usage_5h?.toFixed(4) || '0.0000' }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ sign }}{{ selectedKey.rate_limit_5h?.toFixed(2) || '0.00' }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h ? 'bg-red-500' :
                      selectedKey.usage_5h >= selectedKey.rate_limit_5h * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_5h / selectedKey.rate_limit_5h) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Daily Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit1d') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ sign }}</span>
                <input
                  v-model.number="formData.rate_limit_1d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_1d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'text-red-500' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ sign }}{{ selectedKey.usage_1d?.toFixed(4) || '0.0000' }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ sign }}{{ selectedKey.rate_limit_1d?.toFixed(2) || '0.00' }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d ? 'bg-red-500' :
                      selectedKey.usage_1d >= selectedKey.rate_limit_1d * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_1d / selectedKey.rate_limit_1d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- 7-Day Limit -->
            <div>
              <label class="input-label">{{ t('keys.rateLimit7d') }}</label>
              <div class="relative">
                <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">{{ sign }}</span>
                <input
                  v-model.number="formData.rate_limit_7d"
                  type="number"
                  step="0.01"
                  min="0"
                  class="input pl-7"
                  :placeholder="'0'"
                />
              </div>
              <!-- Usage info (edit mode only) -->
              <div v-if="showEditModal && selectedKey && selectedKey.rate_limit_7d > 0" class="mt-2">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-lg bg-gray-100 px-3 py-2 dark:bg-dark-700 text-sm">
                    <span :class="[
                      'font-medium',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'text-red-500' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'text-yellow-500' :
                      'text-gray-900 dark:text-white'
                    ]">
                      {{ sign }}{{ selectedKey.usage_7d?.toFixed(4) || '0.0000' }}
                    </span>
                    <span class="mx-2 text-gray-400">/</span>
                    <span class="text-gray-500 dark:text-gray-400">
                      {{ sign }}{{ selectedKey.rate_limit_7d?.toFixed(2) || '0.00' }}
                    </span>
                  </div>
                </div>
                <div class="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                  <div
                    :class="[
                      'h-full rounded-full transition-all',
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d ? 'bg-red-500' :
                      selectedKey.usage_7d >= selectedKey.rate_limit_7d * 0.8 ? 'bg-yellow-500' :
                      'bg-green-500'
                    ]"
                    :style="{ width: Math.min((selectedKey.usage_7d / selectedKey.rate_limit_7d) * 100, 100) + '%' }"
                  />
                </div>
              </div>
            </div>

            <!-- Reset Rate Limit button (edit mode only) -->
            <div v-if="showEditModal && selectedKey && (selectedKey.rate_limit_5h > 0 || selectedKey.rate_limit_1d > 0 || selectedKey.rate_limit_7d > 0)">
              <button
                type="button"
                @click="confirmResetRateLimit"
                class="btn btn-secondary text-sm"
              >
                {{ t('keys.resetRateLimitUsage') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Expiration Section -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <label class="input-label mb-0">{{ t('keys.expiration') }}</label>
            <button
              type="button"
              @click="formData.enable_expiration = !formData.enable_expiration"
              :class="[
                'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none',
                formData.enable_expiration ? 'bg-primary-600' : 'bg-gray-200 dark:bg-dark-600'
              ]"
            >
              <span
                :class="[
                  'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                  formData.enable_expiration ? 'translate-x-4' : 'translate-x-0'
                ]"
              />
            </button>
          </div>

          <div v-if="formData.enable_expiration" class="space-y-4 pt-2">
            <!-- Quick select buttons (for both create and edit mode) -->
            <div class="flex flex-wrap gap-2">
              <button
                v-for="days in ['7', '30', '90']"
                :key="days"
                type="button"
                @click="setExpirationDays(parseInt(days))"
                :class="[
                  'rounded-lg px-3 py-1.5 text-sm transition-colors',
                  formData.expiration_preset === days
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
                ]"
              >
                {{ showEditModal ? t('keys.extendDays', { days }) : t('keys.expiresInDays', { days }) }}
              </button>
              <button
                type="button"
                @click="formData.expiration_preset = 'custom'"
                :class="[
                  'rounded-lg px-3 py-1.5 text-sm transition-colors',
                  formData.expiration_preset === 'custom'
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-400'
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600'
                ]"
              >
                {{ t('keys.customDate') }}
              </button>
            </div>

            <!-- Date picker (always show for precise adjustment) -->
            <div>
              <label class="input-label">{{ t('keys.expirationDate') }}</label>
              <input
                v-model="formData.expiration_date"
                type="datetime-local"
                class="input"
              />
              <p class="input-hint">{{ t('keys.expirationDateHint') }}</p>
            </div>

            <!-- Current expiration display (only in edit mode) -->
            <div v-if="showEditModal && selectedKey?.expires_at" class="text-sm">
              <span class="text-gray-500 dark:text-gray-400">{{ t('keys.currentExpiration') }}: </span>
              <span class="font-medium text-gray-900 dark:text-white">
                {{ formatDateTime(selectedKey.expires_at) }}
              </span>
            </div>
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeModals" type="button" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button
            form="key-form"
            type="submit"
            :disabled="submitting"
            class="btn btn-primary"
            data-tour="key-form-submit"
          >
            <svg
              v-if="submitting"
              class="-ml-1 mr-2 h-4 w-4 animate-spin"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              ></circle>
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              ></path>
            </svg>
            {{
              submitting
                ? t('keys.saving')
                : showEditModal
                  ? t('common.update')
                  : t('common.create')
            }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <BulkEditKeysModal
      :show="showBulkEditModal"
      :selected-keys="selectedApiKeys"
      :groups="groups"
      @close="showBulkEditModal = false"
      @updated="handleBulkUpdated"
    />

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('keys.deleteKey')"
      :message="t('keys.deleteConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Reset Quota Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetQuotaDialog"
      :title="t('keys.resetQuotaTitle')"
      :message="t('keys.resetQuotaConfirmMessage', { name: selectedKey?.name, used: selectedKey?.quota_used?.toFixed(4) })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetQuotaUsed"
      @cancel="showResetQuotaDialog = false"
    />

    <!-- Reset Rate Limit Confirmation Dialog -->
    <ConfirmDialog
      :show="showResetRateLimitDialog"
      :title="t('keys.resetRateLimitTitle')"
      :message="t('keys.resetRateLimitConfirmMessage', { name: selectedKey?.name })"
      :confirm-text="t('keys.reset')"
      :cancel-text="t('common.cancel')"
      :danger="true"
      @confirm="resetRateLimitUsage"
      @cancel="showResetRateLimitDialog = false"
    />

    <!-- Use Key Modal -->
    <UseKeyModal
      :show="showUseKeyModal"
      :api-key="selectedKey?.key || ''"
      :base-url="publicSettings?.api_base_url || ''"
      :platform="selectedKey?.group?.platform || null"
      :claude-code-only="selectedKey?.group?.claude_code_only || false"
      :allow-messages-dispatch="selectedKey?.group?.allow_messages_dispatch || false"
      @close="closeUseKeyModal"
    />

    <!-- Less used key actions, teleported so the table's scroll area cannot clip them -->
    <Teleport to="body">
      <div
        v-if="menuKey && menuPosition"
        role="menu"
        class="action-menu-content fixed z-[9999] w-44 overflow-hidden rounded-xl bg-white py-1 shadow-lg ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
        :style="{ top: menuPosition.top + 'px', left: menuPosition.left + 'px' }"
      >
        <router-link
          role="menuitem"
          data-test="key-guide"
          to="/help/guide"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="closeActionMenu"
        >
          <Icon name="book" size="sm" />
          {{ t('yc.connect.toolGuide') }}
        </router-link>
        <button
          role="menuitem"
          data-test="key-toggle"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="runMenuAction(toggleKeyStatus)"
        >
          <Icon :name="menuKey.status === 'active' ? 'ban' : 'checkCircle'" size="sm" />
          {{ menuKey.status === 'active' ? t('keys.disable') : t('keys.enable') }}
        </button>
        <button
          role="menuitem"
          data-test="key-edit"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="runMenuAction(editKey)"
        >
          <Icon name="edit" size="sm" />
          {{ t('common.edit') }}
        </button>
        <div class="my-1 border-t border-gray-100 dark:border-dark-700"></div>
        <button
          role="menuitem"
          data-test="key-delete"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
          @click="runMenuAction(confirmDelete)"
        >
          <Icon name="trash" size="sm" />
          {{ t('common.delete') }}
        </button>
        <div class="my-1 border-t border-gray-100 dark:border-dark-700"></div>
        <button
          role="menuitem"
          data-test="key-manual"
          class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
          @click="runMenuAction(openUseKeyModal)"
        >
          <Icon name="terminal" size="sm" />
          {{ t('keys.useKeyAdvanced') }}
        </button>
      </div>
    </Teleport>

    <!-- CCS Client Selection Dialog for Antigravity -->
    <BaseDialog
      :show="showCcsClientSelect"
      :title="t('keys.ccsClientSelect.title')"
      width="narrow"
      @close="closeCcsClientSelect"
    >
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('keys.ccsClientSelect.description') }}
	        </p>
	        <div class="grid grid-cols-2 gap-3">
	          <button
	            @click="handleCcsClientSelect('claude')"
	            class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 border-gray-200 dark:border-dark-600 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-all"
	          >
	            <Icon name="terminal" size="xl" class="text-gray-600 dark:text-gray-400" />
	            <span class="font-medium text-gray-900 dark:text-white">{{
	              t('keys.ccsClientSelect.claudeCode')
	            }}</span>
	            <span class="text-xs text-gray-500 dark:text-gray-400">{{
	              t('keys.ccsClientSelect.claudeCodeDesc')
	            }}</span>
	          </button>
	          <button
	            @click="handleCcsClientSelect('gemini')"
	            class="flex flex-col items-center gap-2 p-4 rounded-xl border-2 border-gray-200 dark:border-dark-600 hover:border-primary-500 dark:hover:border-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20 transition-all"
	          >
	            <Icon name="sparkles" size="xl" class="text-gray-600 dark:text-gray-400" />
	            <span class="font-medium text-gray-900 dark:text-white">{{
	              t('keys.ccsClientSelect.geminiCli')
	            }}</span>
	            <span class="text-xs text-gray-500 dark:text-gray-400">{{
	              t('keys.ccsClientSelect.geminiCliDesc')
	            }}</span>
	          </button>
	        </div>
	      </div>
      <template #footer>
        <div class="flex justify-end">
          <button @click="closeCcsClientSelect" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Group Selector Dropdown (Teleported to body to avoid overflow clipping) -->
    <Teleport to="body">
      <div
        v-if="groupSelectorKeyId !== null && dropdownPosition"
        ref="dropdownRef"
        class="animate-in fade-in slide-in-from-top-2 fixed z-[100000020] w-max max-w-[calc(100vw-16px)] overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 duration-200 sm:min-w-[380px] dark:bg-dark-800 dark:ring-white/10"
        style="pointer-events: auto !important;"
        :style="{
          top: dropdownPosition.top !== undefined ? dropdownPosition.top + 'px' : undefined,
          bottom: dropdownPosition.bottom !== undefined ? dropdownPosition.bottom + 'px' : undefined,
          left: dropdownPosition.left + 'px'
        }"
      >
        <!-- Search box -->
        <div class="border-b border-gray-100 p-2 dark:border-dark-700">
          <div class="relative">
            <svg class="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="groupSearchQuery"
              type="text"
              class="w-full rounded-lg border border-gray-200 bg-gray-50 py-1.5 pl-8 pr-3 text-sm text-gray-900 placeholder-gray-400 outline-none focus:border-primary-300 focus:ring-1 focus:ring-primary-300 dark:border-dark-600 dark:bg-dark-700 dark:text-white dark:placeholder-gray-500 dark:focus:border-primary-600 dark:focus:ring-primary-600"
              :placeholder="t('keys.searchGroup')"
              @click.stop
            />
          </div>
        </div>
        <!-- Group list -->
        <div class="max-h-80 overflow-y-auto p-1.5">
          <button
            v-for="option in filteredGroupOptions"
            :key="option.value ?? 'null'"
            @click="changeGroup(selectedKeyForGroup!, option.value)"
            :class="[
              'flex w-full items-center justify-between rounded-lg px-3 py-2.5 text-sm transition-colors',
              'border-b border-gray-100 last:border-0 dark:border-dark-700',
              selectedKeyForGroup?.group_id === option.value ||
              (!selectedKeyForGroup?.group_id && option.value === null)
                ? 'bg-primary-50 dark:bg-primary-900/20'
                : 'hover:bg-gray-100 dark:hover:bg-dark-700'
            ]"
            :title="option.description || undefined"
          >
            <GroupOptionItem
              :name="option.label"
              :platform="option.platform"
              :subscription-type="option.subscriptionType"
              :rate-multiplier="option.rate"
              :user-rate-multiplier="option.userRate"
              :peak-rate-enabled="option.peakRateEnabled"
              :peak-start="option.peakStart"
              :peak-end="option.peakEnd"
              :peak-rate-multiplier="option.peakRateMultiplier"
              :description="option.description"
              :selected="
                selectedKeyForGroup?.group_id === option.value ||
                (!selectedKeyForGroup?.group_id && option.value === null)
              "
            />
          </button>
          <!-- Empty state when search has no results -->
          <div v-if="filteredGroupOptions.length === 0" class="py-4 text-center text-sm text-gray-400 dark:text-gray-500">
            {{ t('keys.noGroupFound') }}
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
	import { ref, computed, watch, onMounted, onUnmounted, type ComponentPublicInstance } from 'vue'
	import { useI18n } from 'vue-i18n'
	import { useAppStore } from '@/stores/app'
	import { useOnboardingStore } from '@/stores/onboarding'
	import { useClipboard } from '@/composables/useClipboard'
	import { ycSiteName } from '@/utils/ycBrand'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'

const { t } = useI18n()
import { keysAPI, authAPI, usageAPI, userGroupsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import BulkEditKeysModal from '@/components/keys/BulkEditKeysModal.vue'
	import Pagination from '@/components/common/Pagination.vue'
	import BaseDialog from '@/components/common/BaseDialog.vue'
	import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
	import EmptyState from '@/components/common/EmptyState.vue'
	import Select from '@/components/common/Select.vue'
	import SearchInput from '@/components/common/SearchInput.vue'
	import Icon from '@/components/icons/Icon.vue'
	import UseKeyModal from '@/components/keys/UseKeyModal.vue'
	import EndpointPopover from '@/components/keys/EndpointPopover.vue'
	import GroupBadge from '@/components/common/GroupBadge.vue'
	import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
	import type { ApiKey, Group, PublicSettings, SubscriptionType, GroupPlatform, UpdateApiKeyRequest } from '@/types'
import type { BatchApiKeyUsageStats } from '@/api/usage'
import { formatDateTime } from '@/utils/format'
import { maskApiKey } from '@/utils/maskApiKey'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { platformBadgeLightClass } from '@/utils/platformColors'
import { KEY_GROUP_PROVIDERS, KEY_GROUP_PROVIDER_ICONS, getKeyGroupProvider, type KeyGroupProvider } from '@/utils/keyGroupProviders'
import {
  CC_SWITCH_USAGE_SCRIPT,
  buildCcSwitchImportDeeplink,
  type CcSwitchClientType
} from '@/utils/ccswitchImport'
import { ycSign as sign } from '@/composables/ycMoneyState'


// Helper to format date for datetime-local input
const formatDateTimeLocal = (isoDate: string): string => {
  const date = new Date(isoDate)
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

interface GroupOption {
  value: number
  label: string
  description: string | null
  rate: number
  userRate: number | null
  peakRateEnabled: boolean
  peakStart: string
  peakEnd: string
  peakRateMultiplier: number
  subscriptionType: SubscriptionType
  platform: GroupPlatform
}

const appStore = useAppStore()
const onboardingStore = useOnboardingStore()
const { copyToClipboard: clipboardCopy } = useClipboard()

const SORT_CHOICES = [
  ['created_at', 'desc', 'newest'],
  ['created_at', 'asc', 'oldest'],
  ['last_used_at', 'desc', 'lastUsed'],
  ['name', 'asc', 'name'],
  ['group', 'asc', 'line'],
  ['status', 'asc', 'status'],
  ['expires_at', 'asc', 'expires'],
  ['current_concurrency', 'desc', 'concurrency']
] as const

const sortOptions = computed(() =>
  SORT_CHOICES.map(([key, order, label]) => ({ value: `${key}:${order}`, label: t(`keys.sort.${label}`) }))
)

const apiKeys = ref<ApiKey[]>([])
const selectedIds = ref<number[]>([])
const showBulkEditModal = ref(false)
const selectedApiKeys = computed(() => apiKeys.value.filter((key) => selectedIds.value.includes(key.id)))

const handleSelectionChange = (ids: Array<string | number>) => {
  const visibleIds = new Set(apiKeys.value.map((key) => key.id))
  selectedIds.value = [...new Set(ids.map(Number))].filter((id) => visibleIds.has(id))
}

const toggleSelected = (id: number) => {
  selectedIds.value = selectedIds.value.includes(id)
    ? selectedIds.value.filter((item) => item !== id)
    : [...selectedIds.value, id]
}

const selectAllVisible = () => {
  selectedIds.value = apiKeys.value.map((key) => key.id)
}

const handleBulkUpdated = (succeededIds: number[]) => {
  const succeeded = new Set(succeededIds)
  selectedIds.value = selectedIds.value.filter((id) => !succeeded.has(id))
  loadApiKeys()
}

const groups = ref<Group[]>([])
const loading = ref(false)
const submitting = ref(false)
const now = ref(new Date())
let resetTimer: ReturnType<typeof setInterval> | null = null
const usageStats = ref<Record<string, BatchApiKeyUsageStats>>({})
const usageLoading = ref(false)
const userGroupRates = ref<Record<number, number>>({})

const pagination = ref({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})
const sortState = ref({
  sort_by: 'created_at',
  sort_order: 'desc' as 'asc' | 'desc'
})

// Filter state
const filterSearch = ref('')
const filterStatus = ref('')
const filterGroupId = ref<string | number>('')

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const showResetQuotaDialog = ref(false)
const showResetRateLimitDialog = ref(false)
const showUseKeyModal = ref(false)
const showCcsClientSelect = ref(false)
const pendingCcsRow = ref<ApiKey | null>(null)
const selectedKey = ref<ApiKey | null>(null)
const copiedKeyId = ref<number | null>(null)
const groupSelectorKeyId = ref<number | null>(null)
const publicSettings = ref<PublicSettings | null>(null)
const dropdownRef = ref<HTMLElement | null>(null)
const dropdownPosition = ref<{ top?: number; bottom?: number; left: number } | null>(null)
const groupButtonRefs = ref<Map<number, HTMLElement>>(new Map())
let abortController: AbortController | null = null

// Get the currently selected key for group change
const selectedKeyForGroup = computed(() => {
  if (groupSelectorKeyId.value === null) return null
  return apiKeys.value.find((k) => k.id === groupSelectorKeyId.value) || null
})

const setGroupButtonRef = (keyId: number, el: Element | ComponentPublicInstance | null) => {
  if (el instanceof HTMLElement) {
    groupButtonRefs.value.set(keyId, el)
  } else {
    groupButtonRefs.value.delete(keyId)
  }
}

const formData = ref({
  name: '',
  group_id: null as number | null,
  status: 'active' as 'active' | 'inactive',
  use_custom_key: false,
  custom_key: '',
  enable_ip_restriction: false,
  ip_whitelist: '',
  ip_blacklist: '',
  // Quota settings (empty = unlimited)
  enable_quota: false,
  quota: null as number | null,
  // Rate limit settings
  enable_rate_limit: false,
  rate_limit_5h: null as number | null,
  rate_limit_1d: null as number | null,
  rate_limit_7d: null as number | null,
  enable_expiration: false,
  expiration_preset: '30' as '7' | '30' | '90' | 'custom',
  expiration_date: ''
})

// 自定义Key验证
const customKeyError = computed(() => {
  if (!formData.value.use_custom_key || !formData.value.custom_key) {
    return ''
  }
  const key = formData.value.custom_key
  if (key.length < 16) {
    return t('keys.customKeyTooShort')
  }
  // 检查字符：只允许字母、数字、下划线、连字符
  if (!/^[a-zA-Z0-9_-]+$/.test(key)) {
    return t('keys.customKeyInvalidChars')
  }
  return ''
})

const statusOptions = computed(() => [
  { value: 'active', label: t('common.active') },
  { value: 'inactive', label: t('common.inactive') }
])

const shouldSubmitEditStatus = (key: ApiKey, status: 'active' | 'inactive') => {
  if (key.status === 'quota_exhausted' || key.status === 'expired') {
    return status === 'active'
  }
  return true
}

// Filter dropdown options
const groupFilterOptions = computed(() => [
  { value: '', label: t('keys.allGroups') },
  { value: 0, label: t('keys.noGroup') },
  ...groups.value.map((g) => ({ value: g.id, label: g.name }))
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('keys.allStatus') },
  { value: 'active', label: t('keys.status.active') },
  { value: 'inactive', label: t('keys.status.inactive') },
  { value: 'quota_exhausted', label: t('keys.status.quota_exhausted') },
  { value: 'expired', label: t('keys.status.expired') }
])

const onFilterChange = () => {
  selectedIds.value = []
  pagination.value.page = 1
  loadApiKeys()
}

const onGroupFilterChange = (value: string | number | boolean | null) => {
  filterGroupId.value = value as string | number
  onFilterChange()
}

const onStatusFilterChange = (value: string | number | boolean | null) => {
  filterStatus.value = value as string
  onFilterChange()
}

// Convert groups to Select options format with rate multiplier and subscription type
const groupOptions = computed(() =>
  groups.value.map((group) => ({
    value: group.id,
    label: group.name,
    description: group.description,
    rate: group.rate_multiplier,
    userRate: userGroupRates.value[group.id] ?? null,
    peakRateEnabled: group.peak_rate_enabled,
    peakStart: group.peak_start,
    peakEnd: group.peak_end,
    peakRateMultiplier: group.peak_rate_multiplier,
    subscriptionType: group.subscription_type,
    platform: group.platform
  }))
)

const createProvider = ref<KeyGroupProvider>('anthropic')
const createProviderOptions = computed(() => KEY_GROUP_PROVIDERS.map((value) => ({
  value,
  label: t(`keys.providers.${value}`),
  count: groups.value.filter((group) => getKeyGroupProvider(group.platform) === value).length
})))

const formGroupOptions = computed(() => showEditModal.value
  ? groupOptions.value
  : groupOptions.value.filter((group) => getKeyGroupProvider(group.platform) === createProvider.value)
)

const selectCreateProvider = (provider: KeyGroupProvider) => {
  if (createProvider.value === provider) return
  createProvider.value = provider
  formData.value.group_id = null
}

// Also handles groups arriving after the create dialog has already opened.
watch([showCreateModal, createProviderOptions], ([isOpen, providers], [wasOpen]) => {
  if (!isOpen) return
  if (!wasOpen || !providers.some((provider) => provider.value === createProvider.value && provider.count > 0)) {
    selectCreateProvider(providers.find((provider) => provider.count > 0)?.value ?? 'anthropic')
  }
  if (!formGroupOptions.value.some((group) => group.value === formData.value.group_id)) {
    formData.value.group_id = null
  }
})

// Group dropdown search
const groupSearchQuery = ref('')
const filteredGroupOptions = computed(() => {
  const query = groupSearchQuery.value.trim().toLowerCase()
  if (!query) return groupOptions.value
  return groupOptions.value.filter((opt) => {
    return opt.label.toLowerCase().includes(query) ||
      (opt.description && opt.description.toLowerCase().includes(query))
  })
})

const copyToClipboard = async (text: string, keyId: number) => {
  const success = await clipboardCopy(text, t('keys.copied'))
  if (success) {
    copiedKeyId.value = keyId
    setTimeout(() => {
      copiedKeyId.value = null
    }, 800)
  }
}

const isAbortError = (error: unknown) => {
  if (!error || typeof error !== 'object') return false
  const { name, code } = error as { name?: string; code?: string }
  return name === 'AbortError' || code === 'ERR_CANCELED'
}

const loadApiKeys = async () => {
  abortController?.abort()
  const controller = new AbortController()
  abortController = controller
  const { signal } = controller
  loading.value = true
  usageLoading.value = false
  try {
    // Build filters
    const filters: {
      search?: string
      status?: string
      group_id?: number | string
      sort_by?: string
      sort_order?: 'asc' | 'desc'
    } = {}
    if (filterSearch.value) filters.search = filterSearch.value
    if (filterStatus.value) filters.status = filterStatus.value
    if (filterGroupId.value !== '') filters.group_id = filterGroupId.value
    filters.sort_by = sortState.value.sort_by
    filters.sort_order = sortState.value.sort_order

    const response = await keysAPI.list(pagination.value.page, pagination.value.page_size, filters, {
      signal
    })
    if (signal.aborted) return
    apiKeys.value = response.items
    handleSelectionChange(selectedIds.value)
    pagination.value.total = response.total
    pagination.value.pages = response.pages

    // 用量统计要聚合 30 天的 usage_logs，可能比列表慢很多：不阻塞表格渲染，单独加载
    usageLoading.value = response.items.length > 0
    if (response.items.length > 0) {
      void loadUsageStats(response.items.map((k) => k.id), signal)
    }
  } catch (error) {
    if (isAbortError(error)) {
      return
    }
    appStore.showError(t('keys.failedToLoad'))
  } finally {
    if (abortController === controller) {
      loading.value = false
    }
  }
}

const loadUsageStats = async (keyIds: number[], signal: AbortSignal) => {
  try {
    const usageResponse = await usageAPI.getDashboardApiKeysUsage(keyIds, { signal })
    if (signal.aborted) return
    usageStats.value = usageResponse.stats
  } catch (e) {
    if (!isAbortError(e)) {
      console.error('Failed to load usage stats:', e)
    }
  } finally {
    if (!signal.aborted) {
      usageLoading.value = false
    }
  }
}

const loadGroups = async () => {
  try {
    groups.value = await userGroupsAPI.getAvailable()
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

const loadUserGroupRates = async () => {
  try {
    userGroupRates.value = await userGroupsAPI.getUserGroupRates()
  } catch (error) {
    console.error('Failed to load user group rates:', error)
  }
}

const loadPublicSettings = async () => {
  try {
    publicSettings.value = await authAPI.getPublicSettings()
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
}

const openUseKeyModal = (key: ApiKey) => {
  selectedKey.value = key
  showUseKeyModal.value = true
}

const closeUseKeyModal = () => {
  showUseKeyModal.value = false
  selectedKey.value = null
}

const handlePageChange = (page: number) => {
  selectedIds.value = []
  pagination.value.page = page
  loadApiKeys()
}

const handlePageSizeChange = (pageSize: number) => {
  selectedIds.value = []
  pagination.value.page_size = pageSize
  pagination.value.page = 1
  loadApiKeys()
}

const handleSort = (key: string, order: 'asc' | 'desc') => {
  selectedIds.value = []
  sortState.value.sort_by = key
  sortState.value.sort_order = order
  pagination.value.page = 1
  loadApiKeys()
}

const sortValue = computed(() => `${sortState.value.sort_by}:${sortState.value.sort_order}`)

const onSortChange = (value: string | number | boolean | null) => {
  const [key, order] = String(value).split(':')
  handleSort(key, order === 'asc' ? 'asc' : 'desc')
}

// A line's price next to the standard one, in words instead of a multiplier.
const linePriceTag = (row: ApiKey) => {
  if (!row.group) return ''
  const rate = userGroupRates.value[row.group.id] ?? row.group.rate_multiplier
  if (!(rate > 0) || rate === 1) return ''
  const pct = Math.round(Math.abs(1 - rate) * 100)
  return rate < 1
    ? t('keys.lineCheaper', { zhe: parseFloat((rate * 10).toFixed(1)), pct })
    : t('keys.lineDearer', { pct })
}

const editKey = (key: ApiKey) => {
  selectedKey.value = key
  const hasIPRestriction = (key.ip_whitelist?.length > 0) || (key.ip_blacklist?.length > 0)
  const hasExpiration = !!key.expires_at
  formData.value = {
    name: key.name,
    group_id: key.group_id,
    status: key.status === 'quota_exhausted' || key.status === 'expired' ? 'inactive' : key.status,
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: hasIPRestriction,
    ip_whitelist: (key.ip_whitelist || []).join('\n'),
    ip_blacklist: (key.ip_blacklist || []).join('\n'),
    enable_quota: key.quota > 0,
    quota: key.quota > 0 ? key.quota : null,
    enable_rate_limit: (key.rate_limit_5h > 0) || (key.rate_limit_1d > 0) || (key.rate_limit_7d > 0),
    rate_limit_5h: key.rate_limit_5h || null,
    rate_limit_1d: key.rate_limit_1d || null,
    rate_limit_7d: key.rate_limit_7d || null,
    enable_expiration: hasExpiration,
    expiration_preset: 'custom',
    expiration_date: key.expires_at ? formatDateTimeLocal(key.expires_at) : ''
  }
  showEditModal.value = true
}

const toggleKeyStatus = async (key: ApiKey) => {
  const newStatus = key.status === 'active' ? 'inactive' : 'active'
  try {
    await keysAPI.toggleStatus(key.id, newStatus)
    appStore.showSuccess(
      newStatus === 'active' ? t('keys.keyEnabledSuccess') : t('keys.keyDisabledSuccess')
    )
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToUpdateStatus'))
  }
}

const openGroupSelector = (key: ApiKey) => {
  if (groupSelectorKeyId.value === key.id) {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
  } else {
    const buttonEl = groupButtonRefs.value.get(key.id)
    if (buttonEl) {
      const rect = buttonEl.getBoundingClientRect()
      const dropdownEstHeight = 400 // estimated max dropdown height
      const dropdownEstWidth = Math.min(380, window.innerWidth - 16)
      const spaceBelow = window.innerHeight - rect.bottom
      const spaceAbove = rect.top
      // 夹取 left，避免窄屏下浮层超出视口右缘
      const left = Math.max(8, Math.min(rect.left, window.innerWidth - dropdownEstWidth - 8))

      if (spaceBelow < dropdownEstHeight && spaceAbove > spaceBelow) {
        // Not enough space below, pop upward
        dropdownPosition.value = {
          bottom: window.innerHeight - rect.top + 4,
          left
        }
      } else {
        // Default: pop downward
        dropdownPosition.value = {
          top: rect.bottom + 4,
          left
        }
      }
    }
    groupSelectorKeyId.value = key.id
    groupSearchQuery.value = ''
  }
}

const changeGroup = async (key: ApiKey, newGroupId: number | null) => {
  groupSelectorKeyId.value = null
  dropdownPosition.value = null
  if (key.group_id === newGroupId) return

  try {
    await keysAPI.update(key.id, { group_id: newGroupId })
    appStore.showSuccess(t('keys.groupChangedSuccess'))
    loadApiKeys()
  } catch (error) {
    appStore.showError(t('keys.failedToChangeGroup'))
  }
}

const closeGroupSelector = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  // Check if click is inside the dropdown or the trigger button
  if (!target.closest('.group\\/dropdown') && !dropdownRef.value?.contains(target)) {
    groupSelectorKeyId.value = null
    dropdownPosition.value = null
  }
}

const confirmDelete = (key: ApiKey) => {
  selectedKey.value = key
  showDeleteDialog.value = true
}

const handleSubmit = async () => {
  // Validate group_id is required
  if (formData.value.group_id === null) {
    appStore.showError(t('keys.groupRequired'))
    return
  }

  // Validate custom key if enabled
  if (!showEditModal.value && formData.value.use_custom_key) {
    if (!formData.value.custom_key) {
      appStore.showError(t('keys.customKeyRequired'))
      return
    }
    if (customKeyError.value) {
      appStore.showError(customKeyError.value)
      return
    }
  }

  // Parse IP lists only if IP restriction is enabled
  const parseIPList = (text: string): string[] =>
    text.split('\n').map(ip => ip.trim()).filter(ip => ip.length > 0)
  const ipWhitelist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_whitelist) : []
  const ipBlacklist = formData.value.enable_ip_restriction ? parseIPList(formData.value.ip_blacklist) : []

  // Calculate quota value (null/empty/0 = unlimited, stored as 0)
  const quota = formData.value.quota && formData.value.quota > 0 ? formData.value.quota : 0

  // Calculate expiration
  let expiresInDays: number | undefined
  let expiresAt: string | null | undefined
  if (formData.value.enable_expiration && formData.value.expiration_date) {
    if (!showEditModal.value) {
      // Create mode: calculate days from date
      const expDate = new Date(formData.value.expiration_date)
      const now = new Date()
      const diffDays = Math.ceil((expDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      expiresInDays = diffDays > 0 ? diffDays : 1
    } else {
      // Edit mode: use custom date directly
      expiresAt = new Date(formData.value.expiration_date).toISOString()
    }
  } else if (showEditModal.value) {
    // Edit mode: if expiration disabled or date cleared, send empty string to clear
    expiresAt = ''
  }

  // Calculate rate limit values (send 0 when toggle is off)
  const rateLimitData = formData.value.enable_rate_limit ? {
    rate_limit_5h: formData.value.rate_limit_5h && formData.value.rate_limit_5h > 0 ? formData.value.rate_limit_5h : 0,
    rate_limit_1d: formData.value.rate_limit_1d && formData.value.rate_limit_1d > 0 ? formData.value.rate_limit_1d : 0,
    rate_limit_7d: formData.value.rate_limit_7d && formData.value.rate_limit_7d > 0 ? formData.value.rate_limit_7d : 0,
  } : { rate_limit_5h: 0, rate_limit_1d: 0, rate_limit_7d: 0 }

  submitting.value = true
  try {
    if (showEditModal.value && selectedKey.value) {
      const updates: UpdateApiKeyRequest = {
        name: formData.value.name,
        group_id: formData.value.group_id,
        ip_whitelist: ipWhitelist,
        ip_blacklist: ipBlacklist,
        quota: quota,
        expires_at: expiresAt,
        rate_limit_5h: rateLimitData.rate_limit_5h,
        rate_limit_1d: rateLimitData.rate_limit_1d,
        rate_limit_7d: rateLimitData.rate_limit_7d,
      }
      if (shouldSubmitEditStatus(selectedKey.value, formData.value.status)) {
        updates.status = formData.value.status
      }
      await keysAPI.update(selectedKey.value.id, updates)
      appStore.showSuccess(t('keys.keyUpdatedSuccess'))
    } else {
      const customKey = formData.value.use_custom_key ? formData.value.custom_key : undefined
      await keysAPI.create(
        formData.value.name,
        formData.value.group_id,
        customKey,
        ipWhitelist,
        ipBlacklist,
        quota,
        expiresInDays,
        rateLimitData
      )
      appStore.showSuccess(t('keys.keyCreatedSuccess'))
      // Only advance tour if active, on submit step, and creation succeeded
      if (onboardingStore.isCurrentStep('[data-tour="key-form-submit"]')) {
        onboardingStore.nextStep(500)
      }
    }
    closeModals()
    loadApiKeys()
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToSave')
    appStore.showError(errorMsg)
    // Don't advance tour on error
  } finally {
    submitting.value = false
  }
}

/**
 * 处理删除 API Key 的操作
 * 优化：错误处理改进，优先显示后端返回的具体错误消息（如权限不足等），
 * 若后端未返回消息则显示默认的国际化文本
 */
const handleDelete = async () => {
  if (!selectedKey.value) return

  try {
    await keysAPI.delete(selectedKey.value.id)
    appStore.showSuccess(t('keys.keyDeletedSuccess'))
    showDeleteDialog.value = false
    loadApiKeys()
  } catch (error: any) {
    // 优先使用后端返回的错误消息，提供更具体的错误信息给用户
    const errorMsg = error?.message || t('keys.failedToDelete')
    appStore.showError(errorMsg)
  }
}

const closeModals = () => {
  showCreateModal.value = false
  showEditModal.value = false
  selectedKey.value = null
  formData.value = {
    name: '',
    group_id: null,
    status: 'active',
    use_custom_key: false,
    custom_key: '',
    enable_ip_restriction: false,
    ip_whitelist: '',
    ip_blacklist: '',
    enable_quota: false,
    quota: null,
    enable_rate_limit: false,
    rate_limit_5h: null,
    rate_limit_1d: null,
    rate_limit_7d: null,
    enable_expiration: false,
    expiration_preset: '30',
    expiration_date: ''
  }
}

// Show reset quota confirmation dialog
const confirmResetQuota = () => {
  showResetQuotaDialog.value = true
}

// Set expiration date based on quick select days
const setExpirationDays = (days: number) => {
  formData.value.expiration_preset = days.toString() as '7' | '30' | '90'
  const expDate = new Date()
  expDate.setDate(expDate.getDate() + days)
  formData.value.expiration_date = formatDateTimeLocal(expDate.toISOString())
}

// Reset quota used for an API key
const resetQuotaUsed = async () => {
  const key = selectedKey.value
  if (!key) return
  showResetQuotaDialog.value = false
  try {
    const updatedKey = await keysAPI.update(key.id, { reset_quota: true })
    appStore.showSuccess(t('keys.quotaResetSuccess'))
    key.quota_used = updatedKey.quota_used
    if (key.status !== updatedKey.status) {
      key.status = updatedKey.status
      if (selectedKey.value?.id === key.id) {
        formData.value.status = updatedKey.status === 'active' ? 'active' : 'inactive'
      }
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetQuota')
    appStore.showError(errorMsg)
  }
}

// Show reset rate limit confirmation dialog (from edit modal)
const confirmResetRateLimit = () => {
  showResetRateLimitDialog.value = true
}

// Show reset rate limit confirmation dialog (from table row)
const confirmResetRateLimitFromTable = (row: ApiKey) => {
  selectedKey.value = row
  showResetRateLimitDialog.value = true
}

// Reset rate limit usage for an API key
const resetRateLimitUsage = async () => {
  if (!selectedKey.value) return
  showResetRateLimitDialog.value = false
  try {
    await keysAPI.update(selectedKey.value.id, { reset_rate_limit_usage: true })
    appStore.showSuccess(t('keys.rateLimitResetSuccess'))
    // Refresh key data
    await loadApiKeys()
    // Update the editing key with fresh data
    const refreshedKey = apiKeys.value.find(k => k.id === selectedKey.value!.id)
    if (refreshedKey) {
      selectedKey.value = refreshedKey
    }
  } catch (error: any) {
    const errorMsg = error.response?.data?.detail || t('keys.failedToResetRateLimit')
    appStore.showError(errorMsg)
  }
}

// What a tool's settings ask for next to the key; shown without a trailing slash.
const apiAddress = computed(() => (publicSettings.value?.api_base_url || window.location.origin).replace(/\/+$/, ''))

const activeMenuId = ref<number | null>(null)
const menuPosition = ref<{ top: number; left: number } | null>(null)
const menuKey = computed(() => apiKeys.value.find((key) => key.id === activeMenuId.value) ?? null)

const closeActionMenu = () => {
  activeMenuId.value = null
  menuPosition.value = null
}

const openActionMenu = (key: ApiKey, event: MouseEvent) => {
  if (activeMenuId.value === key.id) {
    closeActionMenu()
    return
  }
  const rect = (event.currentTarget as HTMLElement).getBoundingClientRect()
  const menuWidth = 176
  const menuHeight = 212
  const padding = 8
  const below = rect.bottom + 4
  menuPosition.value = {
    left: Math.max(padding, Math.min(rect.right - menuWidth, window.innerWidth - menuWidth - padding)),
    top: below + menuHeight > window.innerHeight - padding ? Math.max(padding, rect.top - menuHeight - 4) : below
  }
  activeMenuId.value = key.id
}

const runMenuAction = (action: (key: ApiKey) => void | Promise<void>) => {
  const key = menuKey.value
  closeActionMenu()
  if (key) void action(key)
}

const closeActionMenuOnOutsideClick = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.action-menu-trigger') && !target.closest('.action-menu-content')) closeActionMenu()
}

const closeActionMenuOnEscape = (event: KeyboardEvent) => {
  if (event.key === 'Escape') closeActionMenu()
}

const importToCcswitch = (row: ApiKey) => {
  const platform = row.group?.platform || 'anthropic'

  // For antigravity platform, show client selection dialog
  if (platform === 'antigravity') {
    pendingCcsRow.value = row
    showCcsClientSelect.value = true
    return
  }

  // For other platforms, execute directly
  executeCcsImport(row, platform === 'gemini' ? 'gemini' : 'claude')
}

const executeCcsImport = (row: ApiKey, clientType: CcSwitchClientType) => {
  const baseUrl = publicSettings.value?.api_base_url || window.location.origin
  const platform = row.group?.platform || 'anthropic'

  const usageScript = CC_SWITCH_USAGE_SCRIPT
  const providerName = ycSiteName(publicSettings.value?.site_name)
  const deeplink = buildCcSwitchImportDeeplink({
    baseUrl,
    platform,
    clientType,
    providerName,
    apiKey: row.key,
    usageScript
  })

  try {
    window.open(deeplink, '_self')

    // Check if the protocol handler worked by detecting if we're still focused
    setTimeout(() => {
      if (document.hasFocus()) {
        // Still focused means the protocol handler likely failed
        appStore.showError(t('keys.ccSwitchNotInstalled'))
      }
    }, 100)
  } catch (error) {
    appStore.showError(t('keys.ccSwitchNotInstalled'))
  }
}

const handleCcsClientSelect = (clientType: CcSwitchClientType) => {
  if (pendingCcsRow.value) {
    executeCcsImport(pendingCcsRow.value, clientType)
  }
  showCcsClientSelect.value = false
  pendingCcsRow.value = null
}

const closeCcsClientSelect = () => {
  showCcsClientSelect.value = false
  pendingCcsRow.value = null
}

function formatResetTime(resetAt: string | null): string {
  if (!resetAt) return ''
  const diff = new Date(resetAt).getTime() - now.value.getTime()
  if (diff <= 0) return t('keys.resetNow')
  const days = Math.floor(diff / 86400000)
  const hours = Math.floor((diff % 86400000) / 3600000)
  const mins = Math.floor((diff % 3600000) / 60000)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

onMounted(() => {
  loadApiKeys()
  loadGroups()
  loadUserGroupRates()
  loadPublicSettings()
  document.addEventListener('click', closeGroupSelector)
  document.addEventListener('click', closeActionMenuOnOutsideClick)
  document.addEventListener('keydown', closeActionMenuOnEscape)
  window.addEventListener('scroll', closeActionMenu, true)
  resetTimer = setInterval(() => { now.value = new Date() }, 60000)
})

onUnmounted(() => {
  document.removeEventListener('click', closeGroupSelector)
  document.removeEventListener('click', closeActionMenuOnOutsideClick)
  document.removeEventListener('keydown', closeActionMenuOnEscape)
  window.removeEventListener('scroll', closeActionMenu, true)
  if (resetTimer) clearInterval(resetTimer)
})
</script>

<style scoped>
.yk { display: flex; flex-direction: column; gap: 16px; max-width: 1040px; color: var(--yc-ink); }
.yk-grow { flex: 1 1 0; }
.yk-mono { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
.yk-num { font-variant-numeric: tabular-nums; font-weight: 650; }
.yk-dim { color: var(--yc-dim); }
.yk-bad { color: #d92d20; }
.yk-link { font-weight: 600; text-decoration: underline; text-decoration-color: var(--yc-lime); text-decoration-thickness: 2px; text-underline-offset: 3px; }

.yk-btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; height: 36px; padding: 0 14px; border-radius: 999px; font-size: 14px; font-weight: 600; white-space: nowrap; transition: transform 0.15s, box-shadow 0.15s, background 0.15s; }
.yk-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.yk-btn:not(:disabled):active { transform: scale(0.97); }
.yk-btn-lime { background: var(--yc-lime); color: var(--yc-on-lime); box-shadow: 0 0 0 1px rgb(11 15 12 / 0.9); }
.yk-btn-lime:not(:disabled):hover { box-shadow: 0 0 0 1px rgb(11 15 12 / 0.9), 0 3px 0 0 rgb(11 15 12 / 0.9); transform: translateY(-1px); }
.yk-btn-ink { background: var(--yc-ink); color: var(--yc-paper); }
.yk-btn-ghost { color: var(--yc-ink); box-shadow: inset 0 0 0 1px var(--yc-line); background: var(--yc-paper); }
.yk-btn-ghost:not(:disabled):hover, .yk-btn-on { background: var(--yc-fill); }
.yk-btn-icon { width: 36px; padding: 0; }

.yk-address { display: flex; flex-wrap: wrap; align-items: center; gap: 10px 12px; padding: 16px 20px; border-radius: 20px; background: var(--yc-ink); color: var(--yc-paper); }
.yk-address-label { display: inline-flex; align-items: center; height: 24px; padding: 0 10px; border-radius: 999px; font-size: 12.5px; font-weight: 650; background: var(--yc-lime); color: var(--yc-on-lime); }
.yk-address-url { font-size: 15px; word-break: break-all; }
.yk-address .yk-btn-ghost { height: 30px; padding: 0 12px; font-size: 13px; background: transparent; color: inherit; box-shadow: inset 0 0 0 1px rgb(255 255 255 / 0.28); }
:global(.dark) .yk-address .yk-btn-ghost { box-shadow: inset 0 0 0 1px rgb(11 15 12 / 0.28); }
.yk-address .yk-btn-ghost:hover { background: rgb(128 128 128 / 0.22); }
.yk-address-hint { flex: 1 1 100%; font-size: 13.5px; opacity: 0.78; }

.yk-tools { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; }
.yk-bulk { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 16px; background: var(--yc-fill); font-size: 14px; }

.yk-list { display: flex; flex-direction: column; gap: 12px; }
.yk-card { padding: 18px 20px; border-radius: 20px; background: var(--yc-paper); border: 1px solid var(--yc-line); transition: border-color 0.15s, box-shadow 0.15s; }
.yk-card:hover { border-color: rgb(128 128 128 / 0.45); }
.yk-card-picked { border-color: var(--yc-ink); box-shadow: 0 0 0 1px var(--yc-ink); }
.yk-card-off .yk-name, .yk-card-off .yk-key-value { color: var(--yc-dim); }
.yk-empty { padding: 36px 20px; }
.yk-skeleton { display: flex; flex-direction: column; gap: 12px; }
.yk-skeleton i { display: block; height: 14px; border-radius: 7px; background: var(--yc-fill); animation: yk-pulse 1.2s ease-in-out infinite; }
.yk-skeleton i:nth-child(1) { width: 30%; height: 18px; }
.yk-skeleton i:nth-child(2) { width: 62%; }
.yk-skeleton i:nth-child(3) { width: 44%; }
@keyframes yk-pulse { 50% { opacity: 0.45; } }

.yk-head { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px; }
.yk-check { width: 16px; height: 16px; border-radius: 5px; accent-color: var(--yc-ink); opacity: 0.45; cursor: pointer; }
.yk-card:hover .yk-check, .yk-check:checked, .yk-check:focus-visible { opacity: 1; }
.yk-name { font-size: 17px; font-weight: 700; letter-spacing: -0.01em; overflow-wrap: anywhere; }
.yk-shield { color: var(--yc-mut); }
.yk-status { display: inline-flex; align-items: center; gap: 6px; height: 24px; padding: 0 10px; border-radius: 999px; font-size: 12.5px; font-weight: 600; background: var(--yc-fill); color: var(--yc-mut); }
.yk-status::before { content: ''; width: 7px; height: 7px; border-radius: 50%; background: currentColor; }
.yk-status-active { color: var(--yc-ok); }
.yk-status-quota_exhausted { color: #b54708; }
.yk-status-expired { color: #d92d20; }

.yk-key { display: flex; flex-wrap: wrap; align-items: center; gap: 10px; margin-top: 14px; padding: 8px 8px 8px 16px; border-radius: 16px; background: var(--yc-fill); }
.yk-key-value { flex: 1 1 160px; font-size: 15px; letter-spacing: 0.02em; }

.yk-facts { display: flex; flex-wrap: wrap; gap: 10px 28px; margin-top: 14px; font-size: 14px; }
.yk-fact { display: flex; flex-wrap: wrap; align-items: center; gap: 4px 8px; min-width: 0; }
.yk-fact dt { color: var(--yc-mut); }
.yk-fact dd + dt { margin-left: 14px; }
.yk-line { display: inline-flex; flex-wrap: wrap; align-items: center; gap: 6px; margin: -4px -8px; padding: 4px 8px; border-radius: 10px; text-align: left; transition: background 0.15s; }
.yk-line:hover { background: var(--yc-fill); }
.yk-line-name { font-weight: 650; overflow-wrap: anywhere; }
.yk-line-change { font-size: 12.5px; color: var(--yc-mut); text-decoration: underline; text-underline-offset: 3px; }
.yk-tag { display: inline-flex; align-items: center; height: 20px; padding: 0 8px; border-radius: 999px; font-size: 12px; font-weight: 600; background: var(--yc-lime); color: var(--yc-on-lime); }
.yk-pending { display: inline-block; width: 48px; height: 12px; border-radius: 6px; background: var(--yc-fill); animation: yk-pulse 1.2s ease-in-out infinite; }
.yk-quota { max-width: 360px; margin-top: 12px; }

.yk-details { margin-top: 12px; border-top: 1px dashed var(--yc-line); padding-top: 10px; }
.yk-details summary { display: inline-flex; align-items: center; gap: 6px; font-size: 13px; font-weight: 600; color: var(--yc-mut); cursor: pointer; list-style: none; }
.yk-details summary::-webkit-details-marker { display: none; }
.yk-details summary::after { content: '+'; font-size: 15px; line-height: 1; }
.yk-details[open] summary::after { content: '−'; }
.yk-details summary:hover { color: var(--yc-ink); }
.yk-detail-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(190px, 1fr)); gap: 12px 24px; margin-top: 12px; font-size: 13.5px; }
.yk-detail-grid dt { color: var(--yc-mut); font-size: 12.5px; }
.yk-detail-grid dd { margin-top: 2px; overflow-wrap: anywhere; }
.yk-detail-wide { grid-column: 1 / -1; }

@media (max-width: 640px) {
  .yk-card { padding: 16px; }
  .yk-key .yk-btn { flex: 1 1 100%; }
  .yk-tools .yk-grow { display: none; }
}
@media (prefers-reduced-motion: reduce) {
  .yk-btn, .yk-card, .yk-line { transition: none; }
  .yk-skeleton i, .yk-pending { animation: none; }
}
</style>

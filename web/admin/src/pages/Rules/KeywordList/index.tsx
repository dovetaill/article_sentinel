import {
  ModalForm,
  ProFormSelect,
  ProFormSwitch,
  ProFormText,
  ProFormTextArea,
  ProTable,
  type ActionType,
  type ProColumns
} from '@ant-design/pro-components';
import { Button, Card, Input, Popconfirm, Select, Space, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import StatusTag from '@/components/StatusTag';
import { listEnabledCategories, type CategoryRecord } from '@/services/categories';
import {
  createKeyword,
  deleteKeyword,
  listKeywords,
  patchKeywordStatus,
  type KeywordMutationInput,
  type KeywordRecord,
  updateKeyword
} from '@/services/keywords';

const { Title, Paragraph } = Typography;

type CategoryOption = {
  label: string;
  value: number;
};

type KeywordFormValues = {
  name: string;
  category_id: number;
  match_type: string;
  risk_level: string;
  suggest_action: string;
  enabled: boolean;
  remark?: string;
  scopes: string[];
};

type KeywordSearchValues = {
  name: string;
  categoryId?: number;
};

const scopeOptions = [
  { label: '标题', value: 'title' },
  { label: '正文', value: 'body' },
  { label: '关键词字段', value: 'keyword' },
  { label: '富标题', value: 'rich_title' }
];

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

function mapCategoryOptions(items: CategoryRecord[]): CategoryOption[] {
  return items.map((item) => ({
    value: item.id,
    label: item.name
  }));
}

export default function KeywordListPage() {
  const actionRef = useRef<ActionType>();
  const [messageApi, contextHolder] = message.useMessage();
  const [searchParams, setSearchParams] = useSearchParams();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingKeyword, setEditingKeyword] = useState<KeywordRecord | null>(null);
  const [categoryOptions, setCategoryOptions] = useState<CategoryOption[]>([]);
  const [draftFilters, setDraftFilters] = useState<KeywordSearchValues>(() => ({
    name: searchParams.get('name') ?? '',
    categoryId: Number(searchParams.get('category_id') || 0) || undefined
  }));

  const currentPage = normalizePage(searchParams.get('page'));
  const submittedKeyword = searchParams.get('name')?.trim() ?? '';
  const categoryIdFromSearch = Number(searchParams.get('category_id') || 0) || undefined;

  useEffect(() => {
    setDraftFilters({
      name: searchParams.get('name') ?? '',
      categoryId: Number(searchParams.get('category_id') || 0) || undefined
    });
  }, [searchParams]);

  useEffect(() => {
    let cancelled = false;

    void listEnabledCategories()
      .then((items) => {
        if (!cancelled) {
          setCategoryOptions(mapCategoryOptions(items));
        }
      })
      .catch(() => {
        if (!cancelled) {
          setCategoryOptions([]);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const initialValues = useMemo<Partial<KeywordFormValues>>(() => {
    if (!editingKeyword) {
      return {
        category_id: categoryIdFromSearch ?? categoryOptions[0]?.value,
        match_type: 'contains',
        risk_level: 'high',
        suggest_action: 'offline',
        enabled: true,
        scopes: ['title']
      };
    }

    return {
      name: editingKeyword.name,
      category_id: editingKeyword.category_id,
      match_type: editingKeyword.match_type,
      risk_level: editingKeyword.risk_level,
      suggest_action: editingKeyword.suggest_action,
      enabled: editingKeyword.enabled,
      remark: editingKeyword.remark,
      scopes: editingKeyword.scopes
    };
  }, [categoryIdFromSearch, categoryOptions, editingKeyword]);

  const columns: ProColumns<KeywordRecord>[] = [
    {
      title: '关键词名称',
      dataIndex: 'name'
    },
    {
      title: '规则分类',
      dataIndex: 'category_name',
      width: 180
    },
    {
      title: '风险等级',
      dataIndex: 'risk_level',
      width: 120,
      render: (_, record) => <StatusTag kind="risk" value={record.risk_level} />
    },
    {
      title: '适用范围',
      dataIndex: 'scopes',
      render: (_, record) => record.scopes.join('、')
    },
    {
      title: '启用状态',
      dataIndex: 'enabled',
      width: 120,
      render: (_, record) => (
        <Tag className="admin-keyword-status-tag" color={record.enabled ? 'success' : 'default'} bordered={false}>
          {record.enabled ? '启用' : '停用'}
        </Tag>
      )
    },
    {
      title: '操作',
      valueType: 'option',
      width: 260,
      render: (_, record) => [
        <Button
          key="edit"
          type="link"
          onClick={() => {
            setEditingKeyword(record);
            setModalOpen(true);
          }}
        >
          编辑规则
        </Button>,
        <Button
          key="toggle"
          type="link"
          onClick={async () => {
            try {
              await patchKeywordStatus(record.id, !record.enabled);
              messageApi.success(`规则已${record.enabled ? '停用' : '启用'}`);
              actionRef.current?.reload?.();
            } catch (error) {
              messageApi.error(error instanceof Error ? error.message : '规则状态更新失败');
            }
          }}
        >
          {record.enabled ? '停用' : '启用'}
        </Button>,
        <Popconfirm
          key="delete"
          rootClassName="admin-light-popconfirm admin-keyword-popconfirm"
          title="删除规则"
          description="删除后将移除该规则及其扫描范围配置。"
          okText="确认删除"
          cancelText="取消"
          onConfirm={async () => {
            try {
              await deleteKeyword(record.id);
              messageApi.success('规则已删除');
              actionRef.current?.reload?.();
            } catch (error) {
              messageApi.error(error instanceof Error ? error.message : '规则删除失败');
            }
          }}
        >
          <Button type="link" danger>
            删除规则
          </Button>
        </Popconfirm>
      ]
    }
  ];

  return (
    <div className="admin-domain-page">
      {contextHolder}
      <div className="admin-domain-page__head">
        <div>
          <Title level={3} className="admin-domain-page__title">
            规则管理
          </Title>
          <Paragraph className="admin-domain-page__desc">
            维护具体检测规则、风险等级和作用范围，供检测任务直接复用。
          </Paragraph>
        </div>
        <Button
          type="primary"
          onClick={() => {
            setEditingKeyword(null);
            setModalOpen(true);
          }}
        >
          新增规则
        </Button>
      </div>

      <Card className="admin-filter-card admin-surface-panel" variant="borderless">
        <div className="admin-filter-bar">
          <div className="admin-filter-bar__controls">
            <Input
              aria-label="关键词名称"
              className="admin-filter-bar__control"
              placeholder="关键词名称"
              value={draftFilters.name}
              onChange={(event) => {
                setDraftFilters((current) => ({
                  ...current,
                  name: event.target.value
                }));
              }}
            />
            <Select
              allowClear
              showSearch
              optionFilterProp="label"
              aria-label="规则分类"
              className="admin-filter-bar__control admin-filter-bar__control--select"
              placeholder="规则分类"
              value={draftFilters.categoryId}
              options={categoryOptions}
              onChange={(value) => {
                setDraftFilters((current) => ({
                  ...current,
                  categoryId: value
                }));
              }}
            />
          </div>
          <Space wrap>
            <Button onClick={() => setSearchParams(new URLSearchParams())}>重置</Button>
            <Button
              type="primary"
              onClick={() => {
                const nextSearchParams = new URLSearchParams();
                const nextKeyword = draftFilters.name.trim();
                const filtersUnchanged = nextKeyword === submittedKeyword && draftFilters.categoryId === categoryIdFromSearch;

                if (nextKeyword) {
                  nextSearchParams.set('name', nextKeyword);
                }

                if (draftFilters.categoryId) {
                  nextSearchParams.set('category_id', String(draftFilters.categoryId));
                }

                if (filtersUnchanged && currentPage > 1) {
                  nextSearchParams.set('page', String(currentPage));
                }

                setSearchParams(nextSearchParams);
              }}
            >
              查询规则
            </Button>
          </Space>
        </div>

        <div className="admin-table-shell admin-surface-panel">
          <ProTable<KeywordRecord>
            rowKey="id"
            actionRef={actionRef}
            columns={columns}
            search={false}
            options={false}
            cardBordered={false}
            headerTitle={false}
            toolBarRender={false}
            params={{
              name: submittedKeyword,
              category_id: categoryIdFromSearch,
              page: currentPage
            }}
            pagination={{
              current: currentPage,
              pageSize: 20,
              showSizeChanger: false,
              onChange: (nextPage) => {
                if (nextPage === currentPage) {
                  return;
                }

                const nextSearchParams = new URLSearchParams(searchParams);
                if (nextPage > 1) {
                  nextSearchParams.set('page', String(nextPage));
                } else {
                  nextSearchParams.delete('page');
                }

                setSearchParams(nextSearchParams);
              }
            }}
            request={async (params) => {
              try {
                const result = await listKeywords({
                  page: Number(params.current ?? params.page ?? currentPage) || currentPage,
                  pageSize: params.pageSize ?? 20,
                  keyword: submittedKeyword || undefined,
                  categoryId: categoryIdFromSearch
                });

                return {
                  data: result.items,
                  success: true,
                  total: result.total
                };
              } catch (error) {
                messageApi.error(error instanceof Error ? error.message : '规则列表加载失败');
                return {
                  data: [],
                  success: true,
                  total: 0
                };
              }
            }}
          />
        </div>
      </Card>

      <ModalForm<KeywordFormValues>
        open={modalOpen}
        title={editingKeyword ? '编辑规则' : '新增规则'}
        initialValues={initialValues}
        modalProps={{
          rootClassName: 'admin-light-modal admin-keyword-modal',
          destroyOnHidden: true,
          onCancel: () => {
            setModalOpen(false);
            setEditingKeyword(null);
          },
          okText: editingKeyword ? '保存修改' : '确认新增',
          cancelText: '取消'
        }}
        onOpenChange={(open) => {
          setModalOpen(open);
          if (!open) {
            setEditingKeyword(null);
          }
        }}
        onFinish={async (values) => {
          const payload: KeywordMutationInput = {
            name: values.name,
            category_id: Number(values.category_id),
            match_type: values.match_type,
            risk_level: values.risk_level,
            suggest_action: values.suggest_action,
            enabled: values.enabled,
            remark: values.remark,
            scopes: values.scopes
          };

          try {
            if (editingKeyword) {
              await updateKeyword(editingKeyword.id, payload);
              messageApi.success('规则已更新');
            } else {
              await createKeyword(payload);
              messageApi.success('规则已新增');
            }

            setModalOpen(false);
            setEditingKeyword(null);
            actionRef.current?.reload?.();
            return true;
          } catch (error) {
            messageApi.error(error instanceof Error ? error.message : '规则保存失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="关键词名称" fieldProps={{ 'aria-label': '关键词名称' }} rules={[{ required: true }]} />
        <ProFormSelect
          name="category_id"
          label="规则分类"
          fieldProps={{
            options: categoryOptions,
            showSearch: true,
            optionFilterProp: 'label',
            'aria-label': '规则分类'
          }}
          rules={[{ required: true }]}
        />
        <ProFormSelect
          name="match_type"
          label="匹配方式"
          options={[
            { label: '包含匹配', value: 'contains' },
            { label: '完全匹配', value: 'exact' },
            { label: '正则匹配', value: 'regex' }
          ]}
          rules={[{ required: true }]}
        />
        <div className="admin-form-grid admin-form-grid--two-col">
          <ProFormSelect
            name="risk_level"
            label="风险等级"
            options={[
              { label: '低风险', value: 'low' },
              { label: '中风险', value: 'medium' },
              { label: '高风险', value: 'high' }
            ]}
            rules={[{ required: true }]}
          />
          <ProFormSelect
            name="suggest_action"
            label="建议处置"
            options={[
              { label: '忽略', value: 'ignore' },
              { label: '人工处理', value: 'process' },
              { label: '下线处置', value: 'offline' }
            ]}
            rules={[{ required: true }]}
          />
        </div>
        <ProFormSelect
          name="scopes"
          label="作用范围"
          fieldProps={{
            mode: 'multiple',
            options: scopeOptions,
            'aria-label': '作用范围'
          }}
          rules={[{ required: true }]}
        />
        <ProFormTextArea name="remark" label="备注" fieldProps={{ 'aria-label': '备注', rows: 3 }} />
        <ProFormSwitch name="enabled" label="启用状态" />
      </ModalForm>
    </div>
  );
}

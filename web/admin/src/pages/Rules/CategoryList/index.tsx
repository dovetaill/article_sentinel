import {
  ModalForm,
  ProFormDigit,
  ProFormSwitch,
  ProFormText,
  ProTable,
  type ActionType,
  type ProColumns
} from '@ant-design/pro-components';
import { Button, Card, Input, Popconfirm, Select, Space, Tag, Typography, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

import {
  createCategory,
  deleteCategory,
  listCategories,
  patchCategoryStatus,
  type CategoryMutationInput,
  type CategoryRecord,
  updateCategory
} from '@/services/categories';

const { Title, Paragraph } = Typography;

type CategorySearchValues = {
  name: string;
  enabled?: 'true' | 'false';
};

type CategoryFormValues = {
  name: string;
  enabled: boolean;
  sort?: number;
};

function normalizePage(value: string | null) {
  const page = Number(value || 0);
  return Number.isInteger(page) && page > 0 ? page : 1;
}

export default function CategoryListPage() {
  const navigate = useNavigate();
  const actionRef = useRef<ActionType>();
  const [messageApi, contextHolder] = message.useMessage();
  const [searchParams, setSearchParams] = useSearchParams();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryRecord | null>(null);
  const [draftFilters, setDraftFilters] = useState<CategorySearchValues>(() => ({
    name: searchParams.get('name') ?? '',
    enabled: (searchParams.get('enabled') as CategorySearchValues['enabled'] | null) ?? undefined
  }));

  const currentPage = normalizePage(searchParams.get('page'));
  const submittedName = searchParams.get('name')?.trim() ?? '';
  const submittedEnabled = (searchParams.get('enabled') as CategorySearchValues['enabled'] | null) ?? undefined;

  useEffect(() => {
    setDraftFilters({
      name: searchParams.get('name') ?? '',
      enabled: (searchParams.get('enabled') as CategorySearchValues['enabled'] | null) ?? undefined
    });
  }, [searchParams]);

  const initialValues = useMemo<CategoryFormValues>(() => {
    if (!editingCategory) {
      return {
        name: '',
        enabled: true,
        sort: 10
      };
    }

    return {
      name: editingCategory.name,
      enabled: editingCategory.enabled,
      sort: editingCategory.sort
    };
  }, [editingCategory]);

  const columns: ProColumns<CategoryRecord>[] = [
    {
      title: '分类名称',
      dataIndex: 'name'
    },
    {
      title: '排序',
      dataIndex: 'sort',
      width: 100
    },
    {
      title: '启用状态',
      dataIndex: 'enabled',
      width: 120,
      render: (_, record) => (
        <Tag className="admin-category-status-tag" color={record.enabled ? 'success' : 'default'} bordered={false}>
          {record.enabled ? '启用' : '停用'}
        </Tag>
      )
    },
    {
      title: '操作',
      valueType: 'option',
      width: 280,
      render: (_, record) => [
        <Button
          key="edit"
          type="link"
          onClick={() => {
            setEditingCategory(record);
            setModalOpen(true);
          }}
        >
          编辑分类
        </Button>,
        <Button
          key="keywords"
          type="link"
          onClick={() => {
            navigate(`/rules/keywords?category_id=${record.id}`);
          }}
        >
          查看规则
        </Button>,
        <Button
          key="toggle"
          type="link"
          onClick={async () => {
            try {
              await patchCategoryStatus(record.id, !record.enabled);
              messageApi.success(`分类已${record.enabled ? '停用' : '启用'}`);
              actionRef.current?.reload?.();
            } catch (error) {
              messageApi.error(error instanceof Error ? error.message : '分类状态更新失败');
            }
          }}
        >
          {record.enabled ? '停用' : '启用'}
        </Button>,
        <Popconfirm
          key="delete"
          rootClassName="admin-light-popconfirm admin-category-popconfirm"
          title="删除分类"
          description="删除后该分类将不可恢复。"
          okText="确认删除"
          cancelText="取消"
          onConfirm={async () => {
            try {
              await deleteCategory(record.id);
              messageApi.success('分类已删除');
              actionRef.current?.reload?.();
            } catch (error) {
              messageApi.error(error instanceof Error ? error.message : '分类删除失败');
            }
          }}
        >
          <Button type="link" danger>
            删除分类
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
            规则分类
          </Title>
          <Paragraph className="admin-domain-page__desc">
            统一维护分类名称、启停状态和排序，为规则管理与任务编排提供稳定基础。
          </Paragraph>
        </div>
        <Button
          type="primary"
          onClick={() => {
            setEditingCategory(null);
            setModalOpen(true);
          }}
        >
          新增分类
        </Button>
      </div>

      <Card className="admin-filter-card admin-surface-panel" variant="borderless">
        <div className="admin-filter-bar">
          <div className="admin-filter-bar__controls">
            <Input
              aria-label="分类名称"
              className="admin-filter-bar__control"
              placeholder="分类名称"
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
              aria-label="启用状态"
              className="admin-filter-bar__control admin-filter-bar__control--select"
              placeholder="启用状态"
              value={draftFilters.enabled}
              options={[
                { label: '启用', value: 'true' },
                { label: '停用', value: 'false' }
              ]}
              onChange={(value) => {
                setDraftFilters((current) => ({
                  ...current,
                  enabled: value
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
                const nextName = draftFilters.name.trim();
                const filtersUnchanged = nextName === submittedName && draftFilters.enabled === submittedEnabled;

                if (nextName) {
                  nextSearchParams.set('name', nextName);
                }

                if (draftFilters.enabled) {
                  nextSearchParams.set('enabled', draftFilters.enabled);
                }

                if (filtersUnchanged && currentPage > 1) {
                  nextSearchParams.set('page', String(currentPage));
                }

                setSearchParams(nextSearchParams);
              }}
            >
              查询分类
            </Button>
          </Space>
        </div>

        <div className="admin-table-shell admin-surface-panel">
          <ProTable<CategoryRecord>
            rowKey="id"
            actionRef={actionRef}
            columns={columns}
            search={false}
            options={false}
            cardBordered={false}
            headerTitle={false}
            toolBarRender={false}
            params={{
              name: submittedName,
              enabled: submittedEnabled,
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
                const result = await listCategories({
                  page: Number(params.current ?? params.page ?? currentPage) || currentPage,
                  pageSize: params.pageSize ?? 20,
                  name: submittedName || undefined,
                  enabled: submittedEnabled === 'true' ? true : submittedEnabled === 'false' ? false : undefined
                });

                return {
                  data: result.items,
                  success: true,
                  total: result.total
                };
              } catch (error) {
                messageApi.error(error instanceof Error ? error.message : '规则分类加载失败');
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

      <ModalForm<CategoryFormValues>
        open={modalOpen}
        title={editingCategory ? '编辑分类' : '新增分类'}
        initialValues={initialValues}
        modalProps={{
          rootClassName: 'admin-light-modal admin-category-modal',
          destroyOnHidden: true,
          onCancel: () => {
            setModalOpen(false);
            setEditingCategory(null);
          },
          okText: editingCategory ? '保存修改' : '确认新增',
          cancelText: '取消'
        }}
        onOpenChange={(open) => {
          setModalOpen(open);
          if (!open) {
            setEditingCategory(null);
          }
        }}
        onFinish={async (values) => {
          const payload: CategoryMutationInput = {
            name: values.name,
            enabled: values.enabled,
            sort: values.sort ?? 0
          };

          try {
            if (editingCategory) {
              await updateCategory(editingCategory.id, payload);
              messageApi.success('分类已更新');
            } else {
              await createCategory(payload);
              messageApi.success('分类已新增');
            }

            setModalOpen(false);
            setEditingCategory(null);
            actionRef.current?.reload?.();
            return true;
          } catch (error) {
            messageApi.error(error instanceof Error ? error.message : '分类保存失败');
            return false;
          }
        }}
      >
        <ProFormText name="name" label="分类名称" fieldProps={{ 'aria-label': '分类名称' }} rules={[{ required: true }]} />
        <ProFormSwitch name="enabled" label="启用状态" />
        <ProFormDigit name="sort" label="排序" min={0} fieldProps={{ 'aria-label': '排序' }} />
      </ModalForm>
    </div>
  );
}

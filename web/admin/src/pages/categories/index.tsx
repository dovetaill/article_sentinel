import { ModalForm, ProFormDigit, ProFormSwitch, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, Popconfirm, Space, message } from 'antd';
import { useEffect, useMemo, useRef, useState } from 'react';

import { SectionCard } from '../../components/ui/section-card';
import { useOrgContext } from '../../context/org-context';
import {
  createCategory,
  deleteCategory,
  listCategories,
  patchCategoryStatus,
  type CategoryMutationInput,
  type CategoryRecord,
  updateCategory
} from '../../services/categories';

type ActionRef = {
  reload?: () => void;
};

type CategoryFormValues = {
  org_name: string;
  name: string;
  code: string;
  enabled: boolean;
  sort?: number;
};

export default function CategoriesPage() {
  const actionRef = useRef<ActionRef>({});
  const [messageApi, contextHolder] = message.useMessage();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryRecord | null>(null);
  const { activeOrgId, activeOrgName } = useOrgContext();

  useEffect(() => {
    if (activeOrgId) {
      actionRef.current.reload?.();
    }
  }, [activeOrgId]);

  const currentOrgId = activeOrgId ?? 29;
  const currentOrgName = activeOrgName || '一县一端';

  const initialValues = useMemo<CategoryFormValues>(() => {
    if (!editingCategory) {
      return {
        org_name: currentOrgName,
        name: '',
        code: '',
        enabled: true,
        sort: 10
      };
    }

    return {
      org_name: currentOrgName,
      name: editingCategory.name,
      code: editingCategory.code,
      enabled: editingCategory.enabled,
      sort: editingCategory.sort
    };
  }, [currentOrgName, editingCategory]);

  return (
    <>
      {contextHolder}
      <SectionCard
        title="规则分类"
        extra={(
          <Space size={12} wrap>
            <Button
              type="primary"
              onClick={() => {
                setEditingCategory(null);
                setModalOpen(true);
              }}
            >
              新增分类
            </Button>
          </Space>
        )}
      >
        <ProTable<CategoryRecord>
          rowKey="id"
          actionRef={actionRef as never}
          size="small"
          search={{ labelWidth: 88 }}
          cardBordered={false}
          options={false}
          headerTitle={false}
          toolBarRender={false}
          request={async (params) => {
            const enabled = params.enabled === 'true' ? true : params.enabled === 'false' ? false : undefined;
            const result = await listCategories({
              orgid: currentOrgId,
              page: params.current,
              pageSize: params.pageSize,
              name: typeof params.name === 'string' ? params.name : undefined,
              enabled
            });

            return {
              data: result.items,
              success: true,
              total: result.total
            };
          }}
          columns={[
            {
              title: '分类名称',
              dataIndex: 'name'
            },
            {
              title: '分类编码',
              dataIndex: 'code'
            },
            {
              title: '所属机构',
              dataIndex: 'org_name',
              hideInSearch: true,
              render: () => currentOrgName
            },
            {
              title: '排序',
              dataIndex: 'sort',
              hideInSearch: true
            },
            {
              title: '启用状态',
              dataIndex: 'enabled',
              valueType: 'select',
              valueEnum: {
                true: { text: '启用' },
                false: { text: '停用' }
              },
              render: (_, record) => (
                <span className={`status-badge ${record.enabled ? 'status-badge--success' : 'status-badge--neutral'}`}>
                  {record.enabled ? '启用' : '停用'}
                </span>
              )
            },
            {
              title: '操作',
              valueType: 'option',
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
                <Button key="keywords" type="link" href={`/rules/keywords?category_id=${record.id}`}>
                  查看规则
                </Button>,
                <Button
                  key="toggle"
                  type="link"
                  onClick={async () => {
                    await patchCategoryStatus(record.id, currentOrgId, !record.enabled);
                    messageApi.success(`分类已${record.enabled ? '停用' : '启用'}`);
                    actionRef.current.reload?.();
                  }}
                >
                  {record.enabled ? '停用' : '启用'}
                </Button>,
                <Popconfirm
                  key="delete"
                  title="删除分类"
                  description="删除后该分类将不可恢复。"
                  okText="确认删除"
                  cancelText="取消"
                  onConfirm={async () => {
                    await deleteCategory(record.id, currentOrgId);
                    messageApi.success('分类已删除');
                    actionRef.current.reload?.();
                  }}
                >
                  <Button type="link" danger>
                    删除分类
                  </Button>
                </Popconfirm>
              ]
            }
          ]}
        />
      </SectionCard>

      <ModalForm<CategoryFormValues>
        open={modalOpen}
        modalProps={{
          destroyOnHidden: true,
          transitionName: '',
          maskTransitionName: '',
          cancelText: '取消',
          okText: editingCategory ? '保存修改' : '确认新增',
          onCancel: () => {
            setModalOpen(false);
            setEditingCategory(null);
          }
        }}
        title={editingCategory ? '编辑分类' : '新增分类'}
        initialValues={initialValues}
        onFinish={async (values) => {
          const payload: CategoryMutationInput = {
            orgid: currentOrgId,
            name: values.name,
            code: values.code,
            enabled: values.enabled,
            sort: values.sort ?? 0
          };

          if (editingCategory) {
            await updateCategory(editingCategory.id, payload);
            messageApi.success('分类已更新');
          } else {
            await createCategory(payload);
            messageApi.success('分类已新增');
          }

          setModalOpen(false);
          setEditingCategory(null);
          actionRef.current.reload?.();
          return true;
        }}
      >
        <ProFormText name="org_name" label="所属机构" disabled fieldProps={{ 'aria-label': '所属机构' }} />
        <ProFormText name="name" label="分类名称" fieldProps={{ 'aria-label': '分类名称' }} rules={[{ required: true }]} />
        <ProFormText name="code" label="分类编码" fieldProps={{ 'aria-label': '分类编码' }} rules={[{ required: true }]} />
        <ProFormDigit name="sort" label="排序" fieldProps={{ precision: 0, min: 0 }} />
        <ProFormSwitch name="enabled" label="启用状态" />
      </ModalForm>
    </>
  );
}

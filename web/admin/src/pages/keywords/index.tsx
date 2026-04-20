import { ModalForm, ProFormSelect, ProFormSwitch, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, Space, Tag, message } from 'antd';
import { useMemo, useRef, useState } from 'react';

import {
  createKeyword,
  type KeywordMutationInput,
  type KeywordRecord,
  listKeywords,
  updateKeyword
} from '../../services/keywords';

type ActionRef = {
  reload?: () => void;
};

const scopeOptions = [
  { label: 'Title', value: 'title' },
  { label: 'Body', value: 'body' },
  { label: 'Keyword', value: 'keyword' },
  { label: 'Rich Title', value: 'rich_title' }
];

const riskColors: Record<string, string> = {
  low: 'green',
  medium: 'gold',
  high: 'red'
};

export default function KeywordsPage() {
  const actionRef = useRef<ActionRef>({});
  const [messageApi, contextHolder] = message.useMessage();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingKeyword, setEditingKeyword] = useState<KeywordRecord | null>(null);

  const initialValues = useMemo(() => {
    if (!editingKeyword) {
      return {
        orgid: 100,
        match_type: 'contains',
        risk_level: 'high',
        suggest_action: 'offline',
        enabled: true,
        scopes: ['title']
      } satisfies Partial<KeywordMutationInput>;
    }

    return {
      orgid: editingKeyword.orgid,
      name: editingKeyword.name,
      category: editingKeyword.category,
      match_type: editingKeyword.match_type,
      risk_level: editingKeyword.risk_level,
      suggest_action: editingKeyword.suggest_action,
      enabled: editingKeyword.enabled,
      remark: editingKeyword.remark,
      scopes: editingKeyword.scopes
    } satisfies Partial<KeywordMutationInput>;
  }, [editingKeyword]);

  return (
    <>
      {contextHolder}
      <ProTable<KeywordRecord>
        rowKey="id"
        actionRef={actionRef as never}
        search={{ labelWidth: 120 }}
        cardBordered
        headerTitle="Keyword Library"
        toolBarRender={() => [
          <Button
            key="create"
            type="primary"
            onClick={() => {
              setEditingKeyword(null);
              setModalOpen(true);
            }}
          >
            New Keyword
          </Button>
        ]}
        request={async (params) => {
          const result = await listKeywords({
            orgid: 100,
            page: params.current,
            pageSize: params.pageSize,
            category: typeof params.category === 'string' ? params.category : undefined,
            keyword: typeof params.name === 'string' ? params.name : undefined
          });

          return {
            data: result.items,
            success: true,
            total: result.total
          };
        }}
        columns={[
          { title: 'Keyword', dataIndex: 'name' },
          { title: 'Category', dataIndex: 'category' },
          {
            title: 'Risk',
            dataIndex: 'risk_level',
            render: (_, record) => <Tag color={riskColors[record.risk_level] ?? 'default'}>{record.risk_level}</Tag>
          },
          {
            title: 'Scopes',
            dataIndex: 'scopes',
            render: (_, record) => record.scopes.join(', ')
          },
          {
            title: 'Status',
            dataIndex: 'enabled',
            render: (_, record) => <Tag color={record.enabled ? 'green' : 'default'}>{record.enabled ? 'enabled' : 'disabled'}</Tag>
          },
          {
            title: 'Action',
            valueType: 'option',
            render: (_, record) => [
              <Button
                key="edit"
                type="link"
                onClick={() => {
                  setEditingKeyword(record);
                  setModalOpen(true);
                }}
              >
                Edit
              </Button>
            ]
          }
        ]}
      />

      <ModalForm<KeywordMutationInput>
        open={modalOpen}
        modalProps={{
          destroyOnHidden: true,
          transitionName: '',
          maskTransitionName: '',
          cancelText: 'Cancel',
          okText: editingKeyword ? 'Save Changes' : 'Create Keyword',
          onCancel: () => setModalOpen(false)
        }}
        title={editingKeyword ? 'Edit Keyword' : 'Create Keyword'}
        initialValues={initialValues}
        onFinish={async (values) => {
          const payload: KeywordMutationInput = {
            orgid: Number(values.orgid),
            name: values.name,
            category: values.category,
            match_type: values.match_type,
            risk_level: values.risk_level,
            suggest_action: values.suggest_action,
            enabled: values.enabled,
            remark: values.remark,
            scopes: values.scopes
          };

          if (editingKeyword) {
            await updateKeyword(editingKeyword.id, payload);
            messageApi.success('Keyword updated');
          } else {
            await createKeyword(payload);
            messageApi.success('Keyword created');
          }

          setModalOpen(false);
          setEditingKeyword(null);
          actionRef.current.reload?.();
          return true;
        }}
      >
        <ProFormText name="orgid" label="OrgID" disabled />
        <ProFormText name="name" label="Keyword Name" rules={[{ required: true }]} />
        <ProFormText name="category" label="Category" rules={[{ required: true }]} />
        <ProFormSelect
          name="match_type"
          label="Match Type"
          options={[
            { label: 'Contains', value: 'contains' },
            { label: 'Exact', value: 'exact' },
            { label: 'Regex', value: 'regex' }
          ]}
          rules={[{ required: true }]}
        />
        <Space size={16} wrap>
          <ProFormSelect
            name="risk_level"
            label="Risk Level"
            width="sm"
            options={[
              { label: 'Low', value: 'low' },
              { label: 'Medium', value: 'medium' },
              { label: 'High', value: 'high' }
            ]}
            rules={[{ required: true }]}
          />
          <ProFormSelect
            name="suggest_action"
            label="Suggested Action"
            width="sm"
            options={[
              { label: 'Ignore', value: 'ignore' },
              { label: 'Process', value: 'process' },
              { label: 'Offline', value: 'offline' }
            ]}
            rules={[{ required: true }]}
          />
        </Space>
        <ProFormSelect
          name="scopes"
          label="Scopes"
          mode="multiple"
          options={scopeOptions}
          rules={[{ required: true }]}
        />
        <ProFormSwitch name="enabled" label="Enabled" />
        <ProFormText name="remark" label="Remark" />
      </ModalForm>
    </>
  );
}

import { ModalForm, ProFormSelect, ProFormSwitch, ProFormText, ProTable } from '@ant-design/pro-components';
import { Button, Space, message } from 'antd';
import { useMemo, useRef, useState } from 'react';

import { PageHeader } from '../../components/ui/page-header';
import { SectionCard } from '../../components/ui/section-card';
import { StatusBadge } from '../../components/ui/status-badge';
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
  { label: '标题', value: 'title' },
  { label: '正文', value: 'body' },
  { label: '关键词字段', value: 'keyword' },
  { label: '富标题', value: 'rich_title' }
];

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
      <PageHeader
        title="关键词规则"
        description="统一维护巡检词库、匹配方式、风险等级与建议处置，确保巡检口径稳定一致。"
        extra={(
          <Button
            type="primary"
            onClick={() => {
              setEditingKeyword(null);
              setModalOpen(true);
            }}
          >
            新增规则
          </Button>
        )}
      />
      <SectionCard title="规则列表">
        <ProTable<KeywordRecord>
          rowKey="id"
          actionRef={actionRef as never}
          search={{ labelWidth: 110 }}
          cardBordered={false}
          options={{ density: false, fullScreen: false }}
          headerTitle={false}
          toolBarRender={false}
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
            { title: '关键词名称', dataIndex: 'name' },
            { title: '规则分类', dataIndex: 'category' },
            {
              title: '风险等级',
              dataIndex: 'risk_level',
              render: (_, record) => <StatusBadge kind="risk" value={record.risk_level} />
            },
            {
              title: '适用范围',
              dataIndex: 'scopes',
              render: (_, record) => record.scopes.join('、')
            },
            {
              title: '启用状态',
              dataIndex: 'enabled',
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
                    setEditingKeyword(record);
                    setModalOpen(true);
                  }}
                >
                  编辑规则
                </Button>
              ]
            }
          ]}
        />
      </SectionCard>

      <ModalForm<KeywordMutationInput>
        open={modalOpen}
        modalProps={{
          destroyOnHidden: true,
          transitionName: '',
          maskTransitionName: '',
          cancelText: '取消',
          okText: editingKeyword ? '保存修改' : '确认新增',
          onCancel: () => setModalOpen(false)
        }}
        title={editingKeyword ? '编辑规则' : '新增规则'}
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
            messageApi.success('规则已更新');
          } else {
            await createKeyword(payload);
            messageApi.success('规则已新增');
          }

          setModalOpen(false);
          setEditingKeyword(null);
          actionRef.current.reload?.();
          return true;
        }}
      >
        <ProFormText name="orgid" label="所属机构" disabled />
        <ProFormText name="name" label="关键词名称" rules={[{ required: true }]} />
        <ProFormText name="category" label="规则分类" rules={[{ required: true }]} />
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
        <Space size={16} wrap>
          <ProFormSelect
            name="risk_level"
            label="风险等级"
            width="sm"
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
            width="sm"
            options={[
              { label: '忽略', value: 'ignore' },
              { label: '人工处理', value: 'process' },
              { label: '下线处置', value: 'offline' }
            ]}
            rules={[{ required: true }]}
          />
        </Space>
        <ProFormSelect
          name="scopes"
          label="适用范围"
          mode="multiple"
          options={scopeOptions}
          rules={[{ required: true }]}
        />
        <ProFormSwitch name="enabled" label="启用状态" />
        <ProFormText name="remark" label="备注说明" />
      </ModalForm>
    </>
  );
}

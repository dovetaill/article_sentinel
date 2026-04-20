import { ProForm, ProFormText, ProFormTextArea } from '@ant-design/pro-components';
import { Button, Card, Descriptions, Form, Space, Spin, Typography, message } from 'antd';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';

import { getArticleRectify, rectifyArticle, type RectifyArticleRecord } from '../../services/results';

const { Paragraph } = Typography;

type RectifyFormValues = {
  title: string;
  desc: string;
  body: string;
};

export default function RectifyPage() {
  const { articleId } = useParams();
  const [form] = Form.useForm<RectifyFormValues>();
  const [messageApi, contextHolder] = message.useMessage();
  const [loading, setLoading] = useState(true);
  const [record, setRecord] = useState<RectifyArticleRecord | null>(null);
  const [targetState, setTargetState] = useState<number>();

  useEffect(() => {
    if (!articleId) {
      setLoading(false);
      return;
    }

    setLoading(true);
    void getArticleRectify(Number(articleId), 100)
      .then((data) => {
        setRecord(data);
        form.setFieldsValue({
          title: data.title,
          desc: data.desc,
          body: data.body
        });
      })
      .finally(() => setLoading(false));
  }, [articleId, form]);

  return (
    <>
      {contextHolder}
      <Card title={`Rectify Article #${articleId || '-'}`} variant="borderless">
        {loading ? <Spin /> : null}
        {record ? (
          <Space direction="vertical" size={20} style={{ width: '100%' }}>
            <Descriptions
              column={1}
              items={[
                { key: 'title', label: 'Current Title', children: record.title },
                { key: 'desc', label: 'Current Summary', children: record.desc },
                {
                  key: 'body',
                  label: 'Current Body',
                  children: <Paragraph code>{record.body}</Paragraph>
                }
              ]}
            />
            <ProForm<RectifyFormValues>
              form={form}
              submitter={false}
              onFinish={async (values) => {
                if (!articleId) {
                  return false;
                }
                await rectifyArticle(Number(articleId), {
                  orgid: 100,
                  title: values.title,
                  desc: values.desc,
                  body: values.body,
                  target_article_state: targetState
                });
                messageApi.success('Rectification saved');
                setTargetState(undefined);
                return true;
              }}
            >
              <ProFormText
                name="title"
                label="New Title"
                rules={[{ required: true }]}
                fieldProps={{ 'aria-label': 'New Title' }}
              />
              <ProFormText
                name="desc"
                label="New Summary"
                rules={[{ required: true }]}
                fieldProps={{ 'aria-label': 'New Summary' }}
              />
              <ProFormTextArea
                name="body"
                label="New Body HTML"
                rules={[{ required: true }]}
                fieldProps={{ 'aria-label': 'New Body HTML', rows: 8 }}
              />
              <Space>
                <Button type="primary" htmlType="submit" onClick={() => setTargetState(undefined)}>
                  Save Rectification
                </Button>
                <Button htmlType="submit" onClick={() => setTargetState(1)}>
                  Save & Send To Audit
                </Button>
              </Space>
            </ProForm>
          </Space>
        ) : null}
      </Card>
    </>
  );
}

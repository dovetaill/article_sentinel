import { Input, Tabs, Typography } from 'antd';
import { useEffect, useRef, useState } from 'react';

const { Paragraph } = Typography;

type EditorMode = 'visual' | 'source';

export interface HtmlArticleEditorProps {
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
  label?: string;
}

export default function HtmlArticleEditor({
  value = '',
  onChange,
  disabled = false,
  label = 'HTML正文'
}: HtmlArticleEditorProps) {
  const [mode, setMode] = useState<EditorMode>('visual');
  const visualEditorRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const editor = visualEditorRef.current;

    if (!editor) {
      return;
    }

    if (editor.innerHTML !== value) {
      editor.innerHTML = value;
    }
  }, [mode, value]);

  function emitChange(nextValue: string) {
    onChange?.(nextValue);
  }

  return (
    <div className="html-body-editor">
      <Tabs
        activeKey={mode}
        onChange={(nextKey) => setMode(nextKey as EditorMode)}
        items={[
          {
            key: 'visual',
            label: '可视化编辑',
            children: (
              <div className="html-body-editor__panel">
                <Paragraph type="secondary" className="html-body-editor__hint">
                  直接改渲染后的段落内容；需要精细调整标签结构时，可切到 HTML 源码。
                </Paragraph>
                <div
                  ref={visualEditorRef}
                  className={`html-body-editor__canvas${disabled ? ' is-disabled' : ''}`}
                  contentEditable={!disabled}
                  suppressContentEditableWarning
                  role="textbox"
                  aria-label={`${label} 可视化编辑`}
                  aria-multiline="true"
                  data-placeholder="请填写整改正文"
                  onInput={(event) => emitChange(event.currentTarget.innerHTML)}
                  onBlur={(event) => emitChange(event.currentTarget.innerHTML)}
                />
              </div>
            )
          },
          {
            key: 'source',
            label: 'HTML源码',
            children: (
              <div className="html-body-editor__panel">
                <Paragraph type="secondary" className="html-body-editor__hint">
                  粘贴或直接编辑原始 HTML，保存时将以这里的源码为准。
                </Paragraph>
                <Input.TextArea
                  aria-label={`${label} HTML源码`}
                  className="html-body-editor__source"
                  disabled={disabled}
                  rows={12}
                  value={value}
                  onChange={(event) => emitChange(event.target.value)}
                />
              </div>
            )
          }
        ]}
      />
    </div>
  );
}

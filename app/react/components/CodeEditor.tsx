import CodeMirror, {
  keymap,
  oneDarkHighlightStyle,
} from '@uiw/react-codemirror';
import {
  StreamLanguage,
  LanguageSupport,
  syntaxHighlighting,
  indentService,
} from '@codemirror/language';
import { yaml } from '@codemirror/legacy-modes/mode/yaml';
import { dockerFile } from '@codemirror/legacy-modes/mode/dockerfile';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import { useCallback, useMemo, useState } from 'react';
import { createTheme } from '@uiw/codemirror-themes';
import { tags as highlightTags } from '@lezer/highlight';
import type { JSONSchema7 } from 'json-schema';
import { lintKeymap, lintGutter } from '@codemirror/lint';
import { defaultKeymap } from '@codemirror/commands';
import { autocompletion, completionKeymap } from '@codemirror/autocomplete';
import { yamlCompletion, yamlSchema } from 'yaml-schema';

import { AutomationTestingProps } from '@/types';

import { CopyButton } from '@@/buttons/CopyButton';

import { useDebounce } from '../hooks/useDebounce';

import styles from './CodeEditor.module.css';
import { TextTip } from './Tip/TextTip';
import { StackVersionSelector } from './StackVersionSelector';

type Type = 'yaml' | 'shell' | 'dockerfile';
interface Props extends AutomationTestingProps {
  id: string;
  placeholder?: string;
  type?: Type;
  readonly?: boolean;
  onChange?: (value: string) => void;
  value: string;
  height?: string;
  versions?: number[];
  onVersionChange?: (version: number) => void;
  schema?: JSONSchema7;
}

const theme = createTheme({
  theme: 'light',
  settings: {
    background: 'var(--bg-codemirror-color)',
    foreground: 'var(--text-codemirror-color)',
    caret: 'var(--border-codemirror-cursor-color)',
    selection: 'var(--bg-codemirror-selected-color)',
    selectionMatch: 'var(--bg-codemirror-selected-color)',
    gutterBackground: 'var(--bg-codemirror-gutters-color)',
  },
  styles: [
    { tag: highlightTags.atom, color: 'var(--text-cm-default-color)' },
    { tag: highlightTags.meta, color: 'var(--text-cm-meta-color)' },
    {
      tag: [highlightTags.string, highlightTags.special(highlightTags.brace)],
      color: 'var(--text-cm-string-color)',
    },
    { tag: highlightTags.number, color: 'var(--text-cm-number-color)' },
    { tag: highlightTags.keyword, color: 'var(--text-cm-keyword-color)' },
    { tag: highlightTags.comment, color: 'var(--text-cm-comment-color)' },
    {
      tag: highlightTags.variableName,
      color: 'var(--text-cm-variable-name-color)',
    },
  ],
});

// Custom indentation service for YAML
const yamlIndentExtension = indentService.of((context, pos) => {
  const prevLine = context.lineAt(pos, -1);

  // Default to same as previous line
  const prevIndent = /^\s*/.exec(prevLine.text)?.[0].length || 0;

  // If previous line ends with a colon, increase indent
  if (/:\s*$/.test(prevLine.text)) {
    return prevIndent + 2; // Indent 2 spaces after a colon
  }

  return prevIndent;
});

// Create enhanced YAML language with custom indentation (from @codemirror/legacy-modes/mode/yaml)
const yamlLanguageLegacy = new LanguageSupport(StreamLanguage.define(yaml), [
  yamlIndentExtension,
  syntaxHighlighting(oneDarkHighlightStyle),
]);

const dockerFileLanguage = new LanguageSupport(
  StreamLanguage.define(dockerFile)
);
const shellLanguage = new LanguageSupport(StreamLanguage.define(shell));

const docTypeExtensionMap: Record<Type, LanguageSupport> = {
  yaml: yamlLanguageLegacy,
  dockerfile: dockerFileLanguage,
  shell: shellLanguage,
};

function schemaValidationExtensions(schema: JSONSchema7) {
  // skip the hover extension because fields like 'networks' display as 'null' with no description when using the default hover
  // skip the completion extension in favor of custom completion
  const [yaml, linter, , , stateExtensions] = yamlSchema(schema);
  return [
    yaml,
    linter,
    autocompletion({
      icons: false,
      activateOnTypingDelay: 300,
      selectOnOpen: true,
      activateOnTyping: true,
      override: [
        (ctx) => {
          const getCompletions = yamlCompletion();
          const completions = getCompletions(ctx);
          if (Array.isArray(completions)) {
            return null;
          }

          completions.validFor = /^\w*$/;

          return completions;
        },
      ],
    }),
    stateExtensions,
    yamlIndentExtension,
    syntaxHighlighting(oneDarkHighlightStyle),
    lintGutter(),
    keymap.of([...defaultKeymap, ...completionKeymap, ...lintKeymap]),
  ];
}

export function CodeEditor({
  id,
  onChange = () => {},
  placeholder,
  readonly,
  value,
  versions,
  onVersionChange,
  height = '500px',
  type,
  schema,
  'data-cy': dataCy,
}: Props) {
  const [isRollback, setIsRollback] = useState(false);

  const extensions = useMemo(() => {
    if (!type || !docTypeExtensionMap[type]) {
      return [];
    }
    // YAML-specific schema validation
    if (schema && type === 'yaml') {
      return schemaValidationExtensions(schema);
    }
    // Default language support
    return [docTypeExtensionMap[type]];
  }, [type, schema]);

  const handleVersionChange = useCallback(
    (version: number) => {
      if (versions && versions.length > 1) {
        setIsRollback(version < versions[0]);
      }
      onVersionChange?.(version);
    },
    [onVersionChange, versions]
  );

  const [debouncedValue, debouncedOnChange] = useDebounce(value, onChange);

  return (
    <>
      <div className="mb-2 flex flex-col">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            {!!placeholder && <TextTip color="blue">{placeholder}</TextTip>}
          </div>

          <div className="flex-2 ml-auto mr-2 flex items-center gap-x-2">
            <CopyButton
              data-cy={`copy-code-button-${id}`}
              fadeDelay={2500}
              copyText={value}
              color="link"
              className="!pr-0 !text-sm !font-medium hover:no-underline focus:no-underline"
              indicatorPosition="left"
            >
              Copy to clipboard
            </CopyButton>
          </div>
        </div>
        {versions && (
          <div className="mt-2 flex">
            <div className="ml-auto mr-2">
              <StackVersionSelector
                versions={versions}
                onChange={handleVersionChange}
              />
            </div>
          </div>
        )}
      </div>
      <CodeMirror
        className={styles.root}
        theme={theme}
        value={debouncedValue}
        onChange={debouncedOnChange}
        readOnly={readonly || isRollback}
        id={id}
        extensions={extensions}
        height={height}
        basicSetup={{
          highlightSelectionMatches: false,
          autocompletion: !!schema,
        }}
        data-cy={dataCy}
      />
    </>
  );
}

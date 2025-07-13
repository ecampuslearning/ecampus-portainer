import {
  ReactNode,
  ComponentProps,
  PropsWithChildren,
  useMemo,
  useEffect,
} from 'react';
import { useTransitionHook } from '@uirouter/react';
import { JSONSchema7 } from 'json-schema';

import { CodeEditor } from '@@/CodeEditor';

import { FormSectionTitle } from './form-components/FormSectionTitle';
import { FormError } from './form-components/FormError';
import { confirm } from './modals/confirm';
import { ModalType } from './modals';
import { buildConfirmButton } from './modals/utils';
import { ShortcutsTooltip } from './CodeEditor/ShortcutsTooltip';

type CodeEditorProps = ComponentProps<typeof CodeEditor>;

interface Props extends CodeEditorProps {
  titleContent?: ReactNode;
  hideTitle?: boolean;
  error?: string;
  schema?: JSONSchema7;
}

export function WebEditorForm({
  id,
  titleContent = 'Web editor',
  hideTitle,
  children,
  error,
  schema,
  textTip,
  ...props
}: PropsWithChildren<Props>) {
  return (
    <div>
      <div className="web-editor overflow-x-hidden">
        {!hideTitle && (
          <DefaultTitle id={id}>{titleContent ?? null}</DefaultTitle>
        )}
        {children && (
          <div className="form-group text-muted small">
            <div className="col-sm-12 col-lg-12">{children}</div>
          </div>
        )}

        {error && <FormError>{error}</FormError>}

        <div className="form-group">
          <div className="col-sm-12 col-lg-12">
            <CodeEditor
              id={id}
              type="yaml"
              schema={schema as JSONSchema7}
              textTip={textTip}
              // eslint-disable-next-line react/jsx-props-no-spreading
              {...props}
            />
          </div>
        </div>
      </div>
    </div>
  );
}

function DefaultTitle({ id, children }: { id: string; children?: ReactNode }) {
  return (
    <FormSectionTitle htmlFor={id}>
      {children}
      <ShortcutsTooltip />
    </FormSectionTitle>
  );
}

export function usePreventExit(
  initialValue: string,
  value: string,
  check: boolean
) {
  const isChanged = useMemo(
    () => cleanText(initialValue) !== cleanText(value),
    [initialValue, value]
  );

  const preventExit = check && isChanged;

  // when navigating away from the page with unsaved changes, show a portainer prompt to confirm
  useTransitionHook('onBefore', {}, async () => {
    if (!preventExit) {
      return true;
    }
    const confirmed = await confirm({
      modalType: ModalType.Warn,
      title: 'Are you sure?',
      message:
        'You currently have unsaved changes in the text editor. Are you sure you want to leave?',
      confirmButton: buildConfirmButton('Yes', 'danger'),
    });
    return confirmed;
  });

  // when reloading or exiting the page with unsaved changes, show a browser prompt to confirm
  useEffect(() => {
    function handler(event: BeforeUnloadEvent) {
      if (!preventExit) {
        return undefined;
      }

      event.preventDefault();
      // eslint-disable-next-line no-param-reassign
      event.returnValue = '';
      return '';
    }

    // if the form is changed, then set the onbeforeunload
    window.addEventListener('beforeunload', handler);
    return () => {
      window.removeEventListener('beforeunload', handler);
    };
  }, [preventExit]);
}

function cleanText(value: string) {
  return value.replace(/(\r\n|\n|\r)/gm, '');
}

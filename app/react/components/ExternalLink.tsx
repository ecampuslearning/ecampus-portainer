import { ExternalLink as ExternalLinkIcon } from 'lucide-react';
import { PropsWithChildren } from 'react';
import clsx from 'clsx';

import { AutomationTestingProps } from '@/types';

import { Icon } from '@@/Icon';

interface Props {
  to: string;
  className?: string;
}

export function ExternalLink({
  to,
  className,
  children,
  'data-cy': dataCy,
}: PropsWithChildren<Props & AutomationTestingProps>) {
  return (
    <a
      href={to}
      target="_blank"
      rel="noreferrer"
      data-cy={dataCy}
      className={clsx('inline-flex items-center gap-1', className)}
    >
      <Icon icon={ExternalLinkIcon} />
      <span>{children}</span>
    </a>
  );
}

import React from 'react';

import { HelmIcon } from '@/kubernetes/components/helm/helm-templates/HelmIcon';
import { FallbackImage } from '@/react/components/FallbackImage';

import Svg from '@@/Svg';

export interface Chart {
  name: string;
  description: string;
  icon?: string;
  annotations?: {
    category?: string;
  };
}

interface HelmTemplatesListItemProps {
  model: Chart;
  onSelect: (model: Chart) => void;
  actions?: React.ReactNode;
}

export function HelmTemplatesListItem(props: HelmTemplatesListItemProps) {
  const { model, onSelect, actions } = props;

  function handleSelect() {
    onSelect(model);
  }

  return (
    <button
      type="button"
      className="blocklist-item mx-0 bg-inherit text-start"
      onClick={handleSelect}
      tabIndex={0}
    >
      <div className="blocklist-item-box">
        <span className="shrink-0">
          <FallbackImage
            src={model.icon}
            fallbackIcon={HelmIcon}
            className="blocklist-item-logo h-16 w-auto"
            alt="Helm chart icon"
          />
        </span>

        <div className="col-sm-12 flex flex-wrap justify-between gap-2">
          <div className="blocklist-item-line">
            <span>
              <span className="blocklist-item-title">{model.name}</span>
              <span className="space-left blocklist-item-subtitle">
                <span className="vertical-center">
                  <Svg icon="helm" className="icon icon-primary" />
                </span>
                <span> Helm </span>
              </span>
            </span>
          </div>

          <span className="blocklist-item-actions">{actions}</span>

          <div className="blocklist-item-line w-full">
            <span className="blocklist-item-desc">{model.description}</span>
            {model.annotations?.category && (
              <span className="small text-muted">
                {model.annotations.category}
              </span>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}

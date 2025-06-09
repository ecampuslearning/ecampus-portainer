import { useState } from 'react';
import { compact } from 'lodash';

import { useCurrentUser } from '@/react/hooks/useUser';

import { Chart } from '../types';
import { useHelmChartList } from '../queries/useHelmChartList';
import { useHelmRegistries } from '../queries/useHelmRegistries';

import { HelmTemplatesList } from './HelmTemplatesList';
import { HelmTemplatesSelectedItem } from './HelmTemplatesSelectedItem';
import { HelmInstallForm } from './HelmInstallForm';

interface Props {
  onSelectHelmChart: (chartName: string) => void;
  namespace?: string;
  name?: string;
}

export function HelmTemplates({ onSelectHelmChart, namespace, name }: Props) {
  const [selectedChart, setSelectedChart] = useState<Chart | null>(null);
  const [selectedRegistry, setSelectedRegistry] = useState<string | null>(null);

  const { user } = useCurrentUser();
  const helmReposQuery = useHelmRegistries();
  const chartListQuery = useHelmChartList(user.Id, compact([selectedRegistry]));
  function clearHelmChart() {
    setSelectedChart(null);
    onSelectHelmChart('');
  }

  function handleChartSelection(chart: Chart) {
    setSelectedChart(chart);
    onSelectHelmChart(chart.name);
  }

  return (
    <div className="row">
      <div className="col-sm-12 p-0">
        {selectedChart ? (
          <>
            <HelmTemplatesSelectedItem
              selectedChart={selectedChart}
              clearHelmChart={clearHelmChart}
            />
            <HelmInstallForm
              selectedChart={selectedChart}
              namespace={namespace}
              name={name}
            />
          </>
        ) : (
          <HelmTemplatesList
            charts={chartListQuery.data}
            selectAction={handleChartSelection}
            isLoading={chartListQuery.isInitialLoading}
            registries={helmReposQuery.data ?? []}
            selectedRegistry={selectedRegistry}
            setSelectedRegistry={setSelectedRegistry}
          />
        )}
      </div>
    </div>
  );
}
